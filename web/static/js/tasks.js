document.addEventListener('DOMContentLoaded', function() {
    let currentPage = 1;
    const pageSize = 10;
    let totalPages = 1;

    const statusFilter = document.getElementById('statusFilter');
    const taskList = document.getElementById('taskList');
    const prevPageBtn = document.getElementById('prevPage');
    const nextPageBtn = document.getElementById('nextPage');
    const pageInfo = document.getElementById('pageInfo');

    // 加载任务列表
    async function loadTasks() {
        try {
            const response = await fetch(`/api/tasks?page=${currentPage}&size=${pageSize}`);
            const data = await response.json();

            totalPages = Math.ceil(data.total / pageSize);
            updatePagination();

            taskList.innerHTML = data.tasks.map(task => `
                <tr>
                    <td>${task.id}</td>
                    <td>${task.filename}</td>
                    <td>
                        <span class="status-badge ${task.status}">
                            ${getStatusText(task.status)}
                        </span>
                    </td>
                    <td>
                        <div class="progress-bar">
                            <div class="progress-bar-inner" style="width: ${task.progress}%"></div>
                        </div>
                        <span class="progress-text">${task.progress}%</span>
                    </td>
                    <td>${formatDate(task.created_at)}</td>
                    <td>
                        <a href="/task/${task.id}" class="button">
                            <i class="material-icons">visibility</i>
                            查看详情
                        </a>
                    </td>
                </tr>
            `).join('');
        } catch (error) {
            console.error('加载任务列表失败:', error);
        }
    }

    // 更新分页控件
    function updatePagination() {
        prevPageBtn.disabled = currentPage <= 1;
        nextPageBtn.disabled = currentPage >= totalPages;
        pageInfo.textContent = `第 ${currentPage} 页 / 共 ${totalPages} 页`;
    }

    // 状态文本转换
    function getStatusText(status) {
        const statusMap = {
            'processing': '处理中',
            'completed': '已完成',
            'failed': '失败'
        };
        return statusMap[status] || status;
    }

    // 格式化日期
    function formatDate(dateStr) {
        const date = new Date(dateStr);
        return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    }

    // 事件监听
    prevPageBtn.addEventListener('click', () => {
        if (currentPage > 1) {
            currentPage--;
            loadTasks();
        }
    });

    nextPageBtn.addEventListener('click', () => {
        if (currentPage < totalPages) {
            currentPage++;
            loadTasks();
        }
    });

    statusFilter.addEventListener('change', () => {
        currentPage = 1;
        loadTasks();
    });

    // 初始加载
    loadTasks();
}); 