package k8s

import "fmt"

// FormatStorage 格式化存储值，将字节数转换为可读的存储大小表示（如KB, MB, GB等）
func FormatStorage(bytes int64) string {
	if bytes == 0 {
		return "0Gi"
	}

	const (
		kilobyte = 1024
		megabyte = 1024 * kilobyte
		gigabyte = 1024 * megabyte
		terabyte = 1024 * gigabyte
	)

	if bytes >= terabyte {
		return fmt.Sprintf("%.2f Ti", float64(bytes)/float64(terabyte))
	} else if bytes >= gigabyte {
		return fmt.Sprintf("%.2f Gi", float64(bytes)/float64(gigabyte))
	} else if bytes >= megabyte {
		return fmt.Sprintf("%d Mi", bytes/megabyte)
	} else if bytes >= kilobyte {
		return fmt.Sprintf("%d Ki", bytes/kilobyte)
	}
	return fmt.Sprintf("%d B", bytes)
}
