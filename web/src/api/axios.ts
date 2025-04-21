import axios from 'axios';

// create an axios instance
const apiClient = axios.create({
  baseURL: '/api',  // add base URL
  timeout: 5000,    // request timeout
});

// request interceptor
apiClient.interceptors.request.use(
  (config) => {
    // You can add token or other authentication information here
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// response interceptor
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response) {
      // handle error response
      const { status, data } = error.response;
      if (status === 401) {
        // handle unauthorized error  
        console.error('Unauthorized access - perhaps you need to log in?');
        return Promise.reject(new Error('Unauthorized access'));
      } else if (status === 500) {
        // handle server error
        console.error('Server error - please try again later');
        return Promise.reject(new Error('Server error'));
      }
    }
    return Promise.reject(error);
  }
);

const api = apiClient;
export { apiClient, api };
export default api;