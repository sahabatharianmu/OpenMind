import axios from 'axios';

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle 401 Unauthorized
    if (error.response && error.response.status === 401) {
      // Only redirect if we're not already on auth page and we have a token
      // This prevents redirect loops and unnecessary logouts
      const token = localStorage.getItem('access_token');
      if (token && window.location.pathname !== '/auth' && !window.location.pathname.startsWith('/auth')) {
        // Clear tokens and user profile
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user_profile');
        // Use a small delay to allow error handlers to process first
        setTimeout(() => {
          window.location.href = '/auth';
        }, 100);
      }
    }
    return Promise.reject(error);
  }
);

export default api;
