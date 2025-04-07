import React, { useState, useEffect, useRef } from 'react';
import { Card, Select, Button, Space, Input, Switch, Tooltip, message } from 'antd';
import { DownloadOutlined, SyncOutlined, VerticalAlignBottomOutlined } from '@ant-design/icons';
import { getPodLogs, streamPodLogs } from '@/api/pod';
import { useTranslation } from 'react-i18next';

interface PodLogsProps {
  clusterName: string;
  namespace: string;
  podName: string;
  containers: string[];
}

const PodLogs: React.FC<PodLogsProps> = ({ clusterName, namespace, podName, containers }) => {
  const { t } = useTranslation();
  const [selectedContainer, setSelectedContainer] = useState<string>('');
  const [logs, setLogs] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [tailLines, setTailLines] = useState<number>(100);
  const [isStreaming, setIsStreaming] = useState<boolean>(false);
  const [autoScroll, setAutoScroll] = useState<boolean>(true);
  const logsRef = useRef<HTMLDivElement>(null);
  const eventSourceRef = useRef<{ close: () => void } | null>(null);

  // 初始化选择第一个容器
  useEffect(() => {
    if (containers && containers.length > 0 && !selectedContainer) {
      setSelectedContainer(containers[0]);
    }
  }, [containers, selectedContainer]);

  // 当选中的容器变化时，重新获取日志
  useEffect(() => {
    if (selectedContainer) {
      fetchLogs();
    }
    return () => {
      // 清理日志流
      cleanupLogStream();
    };
  }, [selectedContainer]);

  // 监听日志变化，自动滚动到底部
  useEffect(() => {
    if (autoScroll && logsRef.current) {
      logsRef.current.scrollTop = logsRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  // 清理日志流连接
  const cleanupLogStream = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
  };

  // 获取静态日志
  const fetchLogs = async () => {
    if (!selectedContainer) return;
    
    setIsLoading(true);
    setLogs('');
    
    try {
      // 停止之前的流
      cleanupLogStream();
      
      if (isStreaming) {
        // 获取流式日志
        startLogStream();
      } else {
        // 获取静态日志
        const response = await getPodLogs(
          clusterName,
          namespace,
          podName,
          selectedContainer,
          tailLines
        );
        
        if (response.data.data.logs) {
          setLogs(response.data.data.logs);
        } else {
          setLogs(t('podDetail.logs.noData'));
        }
      }
    } catch (error) {
      console.error(t('podDetail.logs.fetchFailed'), error);
      message.error(t('podDetail.logs.fetchFailed'));
      setLogs(t('podDetail.logs.fetchFailedRetry'));
    } finally {
      setIsLoading(false);
    }
  };

  // 开始流式日志
  const startLogStream = () => {
    // 清理旧的日志流
    cleanupLogStream();
    
    // 设置loading状态
    setIsLoading(true);
    
    try {
      // 开始显示连接中的提示
      setLogs(t('podDetail.logs.connectingStream'));
      
      // 开始一个新的日志流
      const logStream = streamPodLogs(
        clusterName,
        namespace,
        podName,
        selectedContainer,
        tailLines,
        true, // follow
        (logLine) => {
          // 接收到日志行时的处理
          setLogs(prevLogs => {
            // 确保日志不会增长得太大，造成性能问题
            const maxLines = 2000;
            const lines = prevLogs.split('\n');
            if (lines.length > maxLines) {
              // 保留最新的日志行，丢弃最旧的
              const truncatedLines = lines.slice(lines.length - maxLines);
              return truncatedLines.join('\n') + '\n' + logLine;
            }
            return prevLogs + logLine + '\n';
          });
          
          // 接收到第一行日志时，结束loading状态
          setIsLoading(false);
        }
      );
      
      // 保存引用以便后续清理
      eventSourceRef.current = logStream;
    } catch (error) {
      console.error(t('podDetail.logs.startStreamFailed'), error);
      message.error(t('podDetail.logs.startStreamFailedNetwork'));
      setLogs(t('podDetail.logs.startStreamFailedDetail', { 
        error: error instanceof Error ? error.message : String(error)
      }));
      setIsLoading(false);
      setIsStreaming(false); // 自动关闭开关
    }
  };

  // 切换日志流模式
  const handleStreamToggle = (checked: boolean) => {
    setIsStreaming(checked);
    if (checked) {
      startLogStream();
    } else {
      cleanupLogStream();
      fetchLogs(); // 切换回静态日志
    }
  };

  // 下载日志
  const downloadLogs = () => {
    if (!logs) return;
    
    const blob = new Blob([logs], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${podName}-${selectedContainer}-logs.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  // 刷新日志
  const refreshLogs = () => {
    fetchLogs();
  };

  return (
    <Card 
      title={t('podDetail.logs.title')}
      extra={
        <Space>
          <Select
            value={selectedContainer}
            onChange={setSelectedContainer}
            style={{ width: 200 }}
            placeholder={t('podDetail.logs.selectContainer')}
            options={containers.map(container => ({ label: container, value: container }))}
          />
          <Input
            type="number"
            addonBefore={t('podDetail.logs.lineCount')}
            value={tailLines}
            onChange={e => setTailLines(parseInt(e.target.value) || 100)}
            style={{ width: 120 }}
          />
          <Tooltip title={t('podDetail.logs.streamTooltip')}>
            <Switch
              checked={isStreaming}
              onChange={handleStreamToggle}
              loading={isLoading}
              checkedChildren={t('podDetail.logs.streamEnabled')}
              unCheckedChildren={t('podDetail.logs.streamDisabled')}
            />
          </Tooltip>
          <Tooltip title={t('podDetail.logs.autoScrollTooltip')}>
            <Switch
              checked={autoScroll}
              onChange={setAutoScroll}
              checkedChildren={t('podDetail.logs.scrollEnabled')}
              unCheckedChildren={t('podDetail.logs.scrollDisabled')}
            />
          </Tooltip>
          <Button 
            icon={<SyncOutlined />} 
            onClick={refreshLogs}
            disabled={isStreaming}
            loading={isLoading}
          >
            {t('common.refresh')}
          </Button>
          <Button 
            icon={<DownloadOutlined />} 
            onClick={downloadLogs}
            disabled={!logs}
          >
            {t('podDetail.logs.download')}
          </Button>
          {!autoScroll && (
            <Button 
              icon={<VerticalAlignBottomOutlined />} 
              onClick={() => {
                if (logsRef.current) {
                  logsRef.current.scrollTop = logsRef.current.scrollHeight;
                }
              }}
            >
              {t('podDetail.logs.scrollToBottom')}
            </Button>
          )}
        </Space>
      }
    >
      <div
        ref={logsRef}
        style={{
          backgroundColor: '#000',
          color: '#fff',
          padding: '10px',
          borderRadius: '4px',
          height: '500px',
          overflow: 'auto',
          fontFamily: 'monospace',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-all'
        }}
      >
        {logs || (isLoading ? t('common.loading') : t('podDetail.logs.noData'))}
      </div>
    </Card>
  );
};

export default PodLogs;