document.addEventListener('DOMContentLoaded', function() {
    const modal = document.getElementById('imagePreview');
    const previewImage = document.getElementById('previewImage');
    const closeBtn = document.querySelector('.close');
    const taskID = window.location.pathname.split('/').pop();

    // 为所有预览按钮添加点击事件
    document.querySelectorAll('.preview-button').forEach(button => {
        button.addEventListener('click', function() {
            const timestamp = this.getAttribute('data-timestamp');
            // 构建图片URL
            const imageUrl = `/api/frames/${taskID}/frame_${timestamp}.jpg`;
            previewImage.src = imageUrl;
            modal.style.display = 'block';
        });
    });

    // 关闭模态框
    closeBtn.addEventListener('click', function() {
        modal.style.display = 'none';
    });

    // 点击模态框外部关闭
    window.addEventListener('click', function(event) {
        if (event.target === modal) {
            modal.style.display = 'none';
        }
    });

    // 键盘事件监听（ESC键关闭模态框）
    document.addEventListener('keydown', function(event) {
        if (event.key === 'Escape' && modal.style.display === 'block') {
            modal.style.display = 'none';
        }
    });
}); 