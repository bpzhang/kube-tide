import axios from 'axios';

// 创建axios实例
const apiClient = axios.create({
  baseURL: '/api',  // 添加基础URL
  timeout: 5000,    // 请求超时时间
});

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    // 在这里可以添加token等认证信息
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response) {
      // 处理错误响应
      const { status, data } = error.response;
      if (status === 401) {
        // 处理未授权错误
      } else if (status === 500) {
        // 处理服务器错误
      }
    }
    return Promise.reject(error);
  }
);

const api = apiClient;
export { apiClient, api };
export default api;