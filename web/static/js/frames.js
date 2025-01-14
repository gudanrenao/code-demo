document.addEventListener('DOMContentLoaded', function() {
    // 标签页切换
    const tabs = document.querySelectorAll('.tab-button');
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            tabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            
            document.querySelectorAll('.tab-content').forEach(content => {
                content.style.display = 'none';
            });
            document.getElementById(`${tab.dataset.tab}-tab`).style.display = 'block';
        });
    });

    // 抽帧类型切换
    const extractTypeInputs = document.querySelectorAll('input[name="extractType"]');
    extractTypeInputs.forEach(input => {
        input.addEventListener('change', () => {
            document.getElementById('interval').disabled = input.value !== 'interval';
            document.getElementById('frameCount').disabled = input.value !== 'count';
        });
    });

    // 抽帧表单提交
    const framesForm = document.getElementById('framesForm');
    framesForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const videoId = this.dataset.videoId;
        if (!videoId) {
            showError('请先上传视频文件');
            return;
        }

        const extractType = document.querySelector('input[name="extractType"]:checked').value;
        const options = {
            interval: extractType === 'interval' ? parseFloat(document.getElementById('interval').value) : undefined,
            frame_count: extractType === 'count' ? parseInt(document.getElementById('frameCount').value) : undefined,
            image_format: document.getElementById('imageFormat').value
        };

        console.log('抽帧选项:', options);

        try {
            const response = await fetch(`/api/video/${videoId}/extract`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(options)
            });

            if (!response.ok) {
                throw new Error('创建抽帧任务失败');
            }

            const result = await response.json();
            showSuccess('抽帧任务已创建');
            pollFrameStatus(result.task_id);
        } catch (error) {
            showError(error.message);
        }
    });

    document.getElementById('videoInput').addEventListener('change', async function(e) {
        const file = e.target.files[0];
        if (!file) return;

        const uploadProgress = document.querySelector('#framesForm .upload-progress');
        uploadProgress.style.display = 'block';

        try {
            // 先上传视频文件
            const formData = new FormData();
            formData.append('file', file);
            
            const uploadResponse = await fetch('/api/upload', {
                method: 'POST',
                body: formData,
            });

            if (!uploadResponse.ok) {
                throw new Error('上传失败');
            }

            const uploadResult = await uploadResponse.json();
            showSuccess('视频上传成功');
            
            // 获取视频信息
            const infoResponse = await fetch(`/api/video/${uploadResult.filename}/info`);
            if (!infoResponse.ok) {
                throw new Error('获取视频信息失败');
            }

            const videoInfo = await infoResponse.json();
            
            // 显示视频信息
            document.getElementById('videoInfo').style.display = 'block';
            document.getElementById('videoDuration').textContent = formatDuration(videoInfo.duration);
            document.getElementById('videoFrameRate').textContent = `${videoInfo.frameRate.toFixed(1)} fps`;
            document.getElementById('videoResolution').textContent = `${videoInfo.width} × ${videoInfo.height}`;
            document.getElementById('videoFormat').textContent = formatVideoFormat(videoInfo.format);
            document.getElementById('videoCodec').textContent = videoInfo.codec;
            document.getElementById('videoBitrate').textContent = formatBitrate(videoInfo.bitrate);
            document.getElementById('videoSize').textContent = formatFileSize(videoInfo.size);
            document.getElementById('audioCodec').textContent = videoInfo.audioCodec || '无';
            document.getElementById('audioChannels').textContent = videoInfo.audioChannels ? 
                `${videoInfo.audioChannels}声道` : '无';
            document.getElementById('audioSampleRate').textContent = formatSampleRate(videoInfo.audioSampleRate);

            // 根据视频时长自动设置合适的抽帧参数
            const interval = document.getElementById('interval');
            const frameCount = document.getElementById('frameCount');
            
            // 设置默认间隔为视频时长的1%，最小0.1秒
            interval.value = Math.max(0.1, (videoInfo.duration * 0.01).toFixed(1));
            // 设置默认帧数为视频总帧数的1%，最小10帧
            frameCount.value = Math.max(10, Math.floor(videoInfo.duration * videoInfo.frameRate * 0.01));

            // 保存文件ID用于后续抽帧
            document.getElementById('framesForm').dataset.videoId = uploadResult.filename;

        } catch (error) {
            showError(error.message);
        } finally {
            uploadProgress.style.display = 'none';
        }
    });
});

function pollFrameStatus(taskId) {
    const interval = setInterval(async () => {
        try {
            const response = await fetch(`/api/tasks/${taskId}`);
            const result = await response.json();
            
            if (result.status === "completed") {
                clearInterval(interval);
                showFrameResults(taskId);
            } else if (result.status === "failed") {
                clearInterval(interval);
                showError('抽帧任务失败');
            }
        } catch (error) {
            console.error('Error polling status:', error);
            clearInterval(interval);
        }
    }, 1000);
}

async function showFrameResults(taskId) {
    try {
        const response = await fetch(`/api/tasks/${taskId}/results`);
        const frames = await response.json();
        
        const resultsDiv = document.getElementById('frameResults');
        const framesGrid = resultsDiv.querySelector('.frames-grid');
        
        // 清空现有内容
        framesGrid.innerHTML = '';
        
        // 处理返回的结果
        const frameList = frames.results || [];
        
        // 按时间戳排序
        const sortedFrames = frameList.sort((a, b) => a.timestamp - b.timestamp);
        
        framesGrid.innerHTML = sortedFrames.map(frame => `
            <div class="frame-item">
                <img src="/api/frames/${taskId}/frame_${frame.timestamp}_time_${frame.videoTime.toFixed(3)}.${frame.imageFormat || 'jpg'}" 
                     alt="Frame" 
                     loading="lazy"
                     data-timestamp="${frame.timestamp}"
                     data-videotime="${frame.videoTime}"
                     ondblclick="showImageModal(this.src)">
                <div class="frame-info">
                    <span>时间: ${formatVideoTimestamp(frame.videoTime)}</span>
                </div>
            </div>
        `).join('');

        // 显示结果区域
        resultsDiv.style.display = 'block';

        // 滚动到结果区域
        resultsDiv.scrollIntoView({ behavior: 'smooth' });

        showSuccess('抽帧完成');
    } catch (error) {
        console.error('获取抽帧结果失败:', error);
        console.error('错误详情:', error.message);
        showError('获取抽帧结果失败');
    }
}

function showError(message) {
    const errorDiv = document.getElementById('errorMessage');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
    setTimeout(() => {
        errorDiv.style.display = 'none';
    }, 5000);
}

function showSuccess(message) {
    const successDiv = document.getElementById('successMessage');
    successDiv.textContent = message;
    successDiv.style.display = 'block';
    setTimeout(() => {
        successDiv.style.display = 'none';
    }, 3000);
}

// 图片预览模态框
const modal = document.getElementById('imageModal');
const modalImg = document.getElementById('modalImage');
const closeButton = document.querySelector('.close-button');

function showImageModal(src) {
    modal.style.display = 'block';
    modalImg.src = src;
}

// 点击关闭按钮关闭模态框
closeButton.onclick = function() {
    modal.style.display = 'none';
}

// 点击模态框背景关闭
modal.onclick = function(e) {
    if (e.target === modal) {
        modal.style.display = 'none';
    }
}

// ESC键关闭模态框
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape' && modal.style.display === 'block') {
        modal.style.display = 'none';
    }
});

// 文件大小格式化函数
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
}

// 时长格式化函数
function formatDuration(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    const ms = Math.floor((seconds % 1) * 1000);

    let result = '';
    if (hours > 0) {
        result += `${hours}:`;
    }
    result += `${minutes.toString().padStart(2, '0')}:`;
    result += `${secs.toString().padStart(2, '0')}`;
    if (ms > 0) {
        result += `.${ms.toString().padStart(3, '0')}`;
    }
    return result;
}

// 视频格式格式化函数
function formatVideoFormat(format) {
    const formatMap = {
        'matroska,webm': 'MKV',
        'mov,mp4,m4a,3gp,3g2,mj2': 'MP4',
        'avi': 'AVI',
        'mpegts': 'TS',
        'flv': 'FLV'
    };
    return formatMap[format] || format.toUpperCase();
}

// 比特率格式化函数
function formatBitrate(bps) {
    if (!bps) return '未知';
    if (bps >= 1000000000) {
        return `${(bps / 1000000000).toFixed(2)} Gbps`;
    } else if (bps >= 1000000) {
        return `${(bps / 1000000).toFixed(2)} Mbps`;
    } else if (bps >= 1000) {
        return `${(bps / 1000).toFixed(2)} Kbps`;
    }
    return `${bps} bps`;
}

// 采样率格式化函数
function formatSampleRate(rate) {
    if (!rate) return '未知';
    if (rate >= 1000) {
        return `${(rate / 1000).toFixed(1)} kHz`;
    }
    return `${rate} Hz`;
}

// 视频时间戳格式化函数 (秒 -> MM:SS 或 HH:MM:SS)
function formatVideoTimestamp(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);

    if (hours > 0) {
        return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    }
    return `${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
} 