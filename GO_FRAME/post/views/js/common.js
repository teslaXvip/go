/* 全局公共方法 */
(function (window) {
    var toastEl = null;

    /* 轻提示：showToast('内容', 'error') */
    function showToast(msg, type) {
        if (!toastEl) {
            toastEl = document.createElement('div');
            toastEl.id = 'toast';
            document.body.appendChild(toastEl);
        }
        toastEl.textContent = msg || '';
        toastEl.className = type === 'error' ? 'show error' : 'show';
        clearTimeout(showToast._timer);
        showToast._timer = setTimeout(function () {
            toastEl.className = '';
        }, 2400);
    }

    /* 表单通用提交：将表单序列化为 urlencoded */
    function formData(form) {
        var body = new URLSearchParams();
        var fd = new FormData(form);
        fd.forEach(function (value, key) {
            body.append(key, value);
        });
        return body;
    }

    window.showToast = showToast;
    window.formData = formData;
})(window);
