// 检查浏览器本地存储的脚本
// 在浏览器控制台中运行以下代码：

console.log("=== 本地存储检查 ===");

// 检查 localStorage
console.log("localStorage 内容:");
for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    const value = localStorage.getItem(key);
    console.log(`${key}:`, value);
}

console.log("\n=== Session Storage 检查 ===");
// 检查 sessionStorage
for (let i = 0; i < sessionStorage.length; i++) {
    const key = sessionStorage.key(i);
    const value = sessionStorage.getItem(key);
    console.log(`${key}:`, value);
}

console.log("\n=== Cookie 检查 ===");
// 检查 cookies
console.log("当前域名的 cookies:");
console.log(document.cookie);

// 获取所有 cookies 的详细信息
const cookies = document.cookie.split(';').map(cookie => {
    const [name, value] = cookie.trim().split('=');
    return { name, value };
});
console.table(cookies);

console.log("\n=== IndexedDB 检查 ===");
// 检查 IndexedDB
if ('indexedDB' in window) {
    indexedDB.databases().then(dbs => {
        console.log("IndexedDB 数据库:", dbs);
    }).catch(err => {
        console.log("无法访问 IndexedDB:", err);
    });
}

console.log("\n=== Cache Storage 检查 ===");
// 检查 Cache Storage
if ('caches' in window) {
    caches.keys().then(cacheNames => {
        console.log("Cache Storage 名称:", cacheNames);
        cacheNames.forEach(name => {
            caches.open(name).then(cache => {
                cache.keys().then(requests => {
                    console.log(`Cache '${name}' 包含 ${requests.length} 个项目`);
                });
            });
        });
    });
}