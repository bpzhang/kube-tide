/**
 * 工具函数模块，提供各种格式化功能
 */

/**
 * 格式化日期时间为本地可读字符串
 * @param dateString 日期字符串或日期对象
 * @param format 可选的格式化选项
 * @returns 格式化后的日期字符串
 */
export function formatDate(dateString: string | Date, format = 'standard'): string {
  if (!dateString) return '-';
  
  const date = new Date(dateString);
  
  if (isNaN(date.getTime())) {
    return '-';
  }
  
  // 针对不同的格式选项提供不同的处理方式
  switch (format) {
    case 'simple':
      return `${date.getFullYear()}-${padZero(date.getMonth() + 1)}-${padZero(date.getDate())}`;
    
    case 'time':
      return `${padZero(date.getHours())}:${padZero(date.getMinutes())}:${padZero(date.getSeconds())}`;
      
    case 'full':
      return `${date.getFullYear()}-${padZero(date.getMonth() + 1)}-${padZero(date.getDate())} ${padZero(date.getHours())}:${padZero(date.getMinutes())}:${padZero(date.getSeconds())}`;
    
    case 'relative':
      return getRelativeTimeString(date);
      
    case 'standard':
    default:
      return `${date.getFullYear()}-${padZero(date.getMonth() + 1)}-${padZero(date.getDate())} ${padZero(date.getHours())}:${padZero(date.getMinutes())}`;
  }
}

/**
 * 将数字填充为两位数 (如 5 -> '05')
 * @param num 要填充的数字
 * @returns 填充后的字符串
 */
function padZero(num: number): string {
  return num.toString().padStart(2, '0');
}

/**
 * 获取相对时间描述
 * @param date 日期对象
 * @returns 相对时间描述
 */
function getRelativeTimeString(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  
  // 转换成秒
  const diffSec = Math.floor(diffMs / 1000);
  
  if (diffSec < 60) {
    return `${diffSec}秒前`;
  }
  
  // 转换成分钟
  const diffMin = Math.floor(diffSec / 60);
  
  if (diffMin < 60) {
    return `${diffMin}分钟前`;
  }
  
  // 转换成小时
  const diffHour = Math.floor(diffMin / 60);
  
  if (diffHour < 24) {
    return `${diffHour}小时前`;
  }
  
  // 转换成天
  const diffDay = Math.floor(diffHour / 24);
  
  if (diffDay < 30) {
    return `${diffDay}天前`;
  }
  
  // 超过30天就显示具体日期
  return formatDate(date, 'standard');
}

/**
 * 格式化文件大小
 * @param bytes 字节数
 * @param decimals 小数位数，默认为2
 * @returns 格式化后的文件大小字符串
 */
export function formatFileSize(bytes: number, decimals: number = 2) {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i];
}

/**
 * 格式化数字，添加千位分隔符
 * @param num 要格式化的数字
 * @returns 格式化后的字符串
 */
export function formatNumber(num: number): string {
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

/**
 * 格式化 CPU 资源，通常以 m（毫核）为单位
 * @param cpu CPU值，如 "100m" 或 0.1
 * @returns 格式化后的CPU值
 */
export function formatCPU(cpu: string | number): string {
  if (!cpu) return '0';
  
  // 如果是字符串形式，如 "100m"
  if (typeof cpu === 'string') {
    if (cpu.endsWith('m')) {
      return cpu;
    }
    
    try {
      // 尝试将字符串转为数字
      const value = parseFloat(cpu);
      return `${value * 1000}m`;
    } catch (e) {
      return cpu;
    }
  }
  
  // 如果是数字形式，如 0.1 (表示0.1核，即100m)
  if (typeof cpu === 'number') {
    return `${cpu * 1000}m`;
  }
  
  return String(cpu);
}

/**
 * 格式化内存资源，通常以 Ki、Mi、Gi 为单位
 * @param memory 内存值，如 "100Mi" 或 104857600
 * @returns 格式化后的内存值
 */
export function formatMemory(memory: string | number): string {
  if (!memory) return '0';
  
  // 如果是字符串形式且已经包含单位，如 "100Mi"
  if (typeof memory === 'string') {
    if (/[KMGTPEkmpgtpe]i?$/i.test(memory)) {
      return memory;
    }
    
    try {
      // 尝试将字符串转为数字
      const value = parseInt(memory, 10);
      return formatBytes(value);
    } catch (e) {
      return memory;
    }
  }
  
  // 如果是数字形式，假设是字节数
  if (typeof memory === 'number') {
    return formatBytes(memory);
  }
  
  return String(memory);
}

/**
 * 将字节数转换为最合适的单位
 * @param bytes 字节数
 * @returns 格式化后的内存字符串
 */
function formatBytes(bytes: number): string {
  const units = ['B', 'Ki', 'Mi', 'Gi', 'Ti', 'Pi', 'Ei'];
  let value = bytes;
  let unitIndex = 0;
  
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }
  
  return `${Math.round(value * 100) / 100} ${units[unitIndex]}`;
}