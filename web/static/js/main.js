document.addEventListener('DOMContentLoaded', function() {
    const uploadForm = document.getElementById('uploadForm');
    const resultContent = document.getElementById('resultContent');
    const errorMessage = document.getElementById('errorMessage');
    const successMessage = document.getElementById('successMessage');
    const uploadProgress = document.querySelector('.upload-progress');
    const progressBar = document.querySelector('.progress-bar-inner');
    const progressText = document.querySelector('.progress-text');

    function showError(message) {
        errorMessage.textContent = message;
        errorMessage.style.display = 'block';
        setTimeout(() => {
            errorMessage.style.display = 'none';
        }, 5000);
    }

    function showSuccess(message) {
        successMessage.textContent = message;
        successMessage.style.display = 'block';
        setTimeout(() => {
            successMessage.style.display = 'none';
        }, 3000);
    }

    function updateProgress(percent) {
        progressBar.style.width = `${percent}%`;
        progressText.textContent = `${Math.round(percent)}%`;
    }

    uploadForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const formData = new FormData(uploadForm);
        uploadProgress.style.display = 'block';
        updateProgress(0);

        try {
            // 上传文件
            const xhr = new XMLHttpRequest();
            xhr.open('POST', '/api/upload', true);

            xhr.upload.onprogress = function(e) {
                if (e.lengthComputable) {
                    const percent = (e.loaded / e.total) * 100;
                    updateProgress(percent);
                }
            };

            xhr.onload = async function() {
                if (xhr.status === 200) {
                    const uploadResult = JSON.parse(xhr.responseText);
                    showSuccess('文件上传成功');

                    // 创建检测任务
                    try {
                        const taskResponse = await fetch('/api/tasks', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json'
                            },
                            body: JSON.stringify({
                                filename: uploadResult.filename,
                                types: ['black_screen', 'color_screen'],
                                params: {
                                    threshold: 0.5
                                }
                            })
                        });

                        if (!taskResponse.ok) {
                            throw new Error('创建任务失败');
                        }

                        const taskResult = await taskResponse.json();
                        showSuccess('任务创建成功');
                        pollTaskStatus(taskResult.task_id);
                    } catch (error) {
                        showError('创建任务失败: ' + error.message);
                    }
                } else {
                    showError('上传失败: ' + xhr.statusText);
                }
                uploadProgress.style.display = 'none';
            };

            xhr.onerror = function() {
                showError('上传失败，请检查网络连接');
                uploadProgress.style.display = 'none';
            };

            xhr.send(formData);
        } catch (error) {
            showError('操作失败: ' + error.message);
            uploadProgress.style.display = 'none';
        }
    });

    async function pollTaskStatus(taskId) {
        const interval = setInterval(async () => {
            try {
                const response = await fetch(`/api/tasks/${taskId}`);
                const result = await response.json();
                
                updateProgress(result.progress);
                
                if (result.status === 'completed') {
                    clearInterval(interval);
                    await showResults(taskId);
                } else if (result.status === 'failed') {
                    clearInterval(interval);
                    alert('检测任务失败');
                }
            } catch (error) {
                console.error('Error polling status:', error);
                clearInterval(interval);
            }
        }, 1000);
    }

    async function showResults(taskId) {
        try {
            const response = await fetch(`/api/tasks/${taskId}/results`);
            const results = await response.json();
            
            let html = '<h3>检测结果:</h3>';
            html += '<ul>';
            for (const result of results.results) {
                html += `<li>
                    类型: ${result.type}<br>
                    时间: ${result.timestamp}s<br>
                    置信度: ${(result.confidence * 100).toFixed(2)}%
                </li>`;
            }
            html += '</ul>';
            
            resultContent.innerHTML = html;
        } catch (error) {
            console.error('Error showing results:', error);
            alert('获取结果失败');
        }
    }
}); 