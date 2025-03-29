import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'
import '@ant-design/v5-patch-for-react-19';

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
