/*
Copyright (c) 2025 Tethys Plex

This file is part of REXI-hub v2.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
import i18n from '../i18n/i18n';
import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import i18next from 'i18next';
import { Label } from 'semantic-ui-react';
import {
  API,
  copy,
  isAdmin,
  showError,
  showSuccess,
  timestamp2string,
} from '../helpers';

import {
  Table,
  Tag,
  Form,
  Button,
  Layout,
  Modal,
  Typography,
  Progress,
  Card,
} from '@douyinfe/semi-ui';
import { ITEMS_PER_PAGE } from '../constants';

const colors = [
  'amber',
  'blue',
  'cyan',
  'green',
  'grey',
  'indigo',
  'light-blue',
  'lime',
  'orange',
  'pink',
  'purple',
  'red',
  'teal',
  'violet',
  'yellow',
];

const renderTimestamp = (timestampInSeconds) => {
  const date = new Date(timestampInSeconds * 1000); // Chuyển từ giây sang mili giây

  const year = date.getFullYear(); // Lấy năm
  const month = ('0' + (date.getMonth() + 1)).slice(-2); // Lấy tháng, bắt đầu từ 0 nên cần +1, và đảm bảo hai chữ số
  const day = ('0' + date.getDate()).slice(-2); // Lấy ngày và đảm bảo hai chữ số
  const hours = ('0' + date.getHours()).slice(-2); // Lấy giờ và đảm bảo hai chữ số
  const minutes = ('0' + date.getMinutes()).slice(-2); // Lấy phút và đảm bảo hai chữ số
  const seconds = ('0' + date.getSeconds()).slice(-2); // Lấy giây và đảm bảo hai chữ số

  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`; // Định dạng đầu ra
};

function renderDuration(submit_time, finishTime) {
  // Đảm bảo startTime và finishTime đều là timestamp hợp lệ
  if (!submit_time || !finishTime) return 'N/A';

  // Chuyển timestamp thành đối tượng Date
  const start = new Date(submit_time);
  const finish = new Date(finishTime);

  // Tính khoảng thời gian (mili giây)
  const durationMs = finish - start;

  // Chuyển khoảng thời gian thành giây, giữ một chữ số thập phân
  const durationSec = (durationMs / 1000).toFixed(1);

  // Đặt màu: lớn hơn 60 giây là đỏ, nhỏ hơn hoặc bằng 60 giây là xanh lá
  const color = durationSec > 60 ? 'red' : 'green';

  // Trả về thẻ màu có định dạng
  return (
    <Tag color={color} size='large'>
      {durationSec} {i18next.t('秒')}
    </Tag>
  );
}

const LogsTable = () => {
  const { t } = useTranslation();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalContent, setModalContent] = useState('');
  const isAdminUser = isAdmin();
  const columns = [
    {
      title: t('提交时间'),
      dataIndex: 'submit_time',
      render: (text, record, index) => {
        return <div>{text ? renderTimestamp(text) : '-'}</div>;
      },
    },
    {
      title: t('结束时间'),
      dataIndex: 'finish_time',
      render: (text, record, index) => {
        return <div>{text ? renderTimestamp(text) : '-'}</div>;
      },
    },
    {
      title: t('进度'),
      dataIndex: 'progress',
      width: 50,
      render: (text, record, index) => {
        return (
          <div>
            {
              // Chuyển ví dụ 100% thành số 100, nếu text không được định nghĩa thì trả về 0
              isNaN(text.replace('%', '')) ? (
                text
              ) : (
                <Progress
                  width={42}
                  type='circle'
                  showInfo={true}
                  percent={Number(text.replace('%', '') || 0)}
                  aria-label={i18n.t('drawing progress')}
                />
              )
            }
          </div>
        );
      },
    },
    {
      title: t('花费时间'),
      dataIndex: 'finish_time', // Dùng finish_time làm dataIndex
      key: 'finish_time',
      render: (finish, record) => {
        // Giả sử record.start_time tồn tại, và finish là timestamp thời điểm hoàn thành
        return <>{finish ? renderDuration(record.submit_time, finish) : '-'}</>;
      },
    },
    {
      title: t('渠道'),
      dataIndex: 'channel_id',
      className: isAdminUser ? 'tableShow' : 'tableHiddle',
      render: (text, record, index) => {
        return (
          <div>
            <Tag
              color={colors[parseInt(text) % colors.length]}
              size='large'
              onClick={() => {
                copyText(text); // Giả sử copyText là hàm dùng để sao chép văn bản
              }}
            >
              {' '}
              {text}{' '}
            </Tag>
          </div>
        );
      },
    },
    {
      title: t('平台'),
      dataIndex: 'platform',
      render: (text, record, index) => {
        return <div>{renderPlatform(text)}</div>;
      },
    },
    {
      title: t('类型'),
      dataIndex: 'action',
      render: (text, record, index) => {
        return <div>{renderType(text)}</div>;
      },
    },
    {
      title: t('任务ID（点击查看详情）'),
      dataIndex: 'task_id',
      render: (text, record, index) => {
        return (
          <Typography.Text
            ellipsis={{ showTooltip: true }}
            //style={{width: 100}}
            onClick={() => {
              setModalContent(JSON.stringify(record, null, 2));
              setIsModalOpen(true);
            }}
          >
            <div>{text}</div>
          </Typography.Text>
        );
      },
    },
    {
      title: t('任务状态'),
      dataIndex: 'status',
      render: (text, record, index) => {
        return <div>{renderStatus(text)}</div>;
      },
    },

    {
      title: t('失败原因'),
      dataIndex: 'fail_reason',
      render: (text, record, index) => {
        // Nếu text không được định nghĩa thì trả về văn bản thay thế, ví dụ chuỗi rỗng '' hoặc giá trị khác
        if (!text) {
          return 'None';
        }

        return (
          <Typography.Text
            ellipsis={{ showTooltip: true }}
            style={{ width: 100 }}
            onClick={() => {
              setModalContent(text);
              setIsModalOpen(true);
            }}
          >
            {text}
          </Typography.Text>
        );
      },
    },
  ];

  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [logCount, setLogCount] = useState(ITEMS_PER_PAGE);
  const [logType] = useState(0);

  let now = new Date();
  // Khởi tạo start_timestamp là ngày hôm trước
  let zeroNow = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const [inputs, setInputs] = useState({
    channel_id: '',
    task_id: '',
    start_timestamp: timestamp2string(zeroNow.getTime() / 1000),
    end_timestamp: '',
  });
  const { channel_id, task_id, start_timestamp, end_timestamp } = inputs;

  const handleInputChange = (value, name) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const setLogsFormat = (logs) => {
    for (let i = 0; i < logs.length; i++) {
      logs[i].timestamp2string = timestamp2string(logs[i].created_at);
      logs[i].key = '' + logs[i].id;
    }
    // data.key = '' + data.id
    setLogs(logs);
    setLogCount(logs.length + ITEMS_PER_PAGE);
    // console.log(logCount);
  };

  const loadLogs = async (startIdx) => {
    setLoading(true);

    let url = '';
    let localStartTimestamp = parseInt(Date.parse(start_timestamp) / 1000);
    let localEndTimestamp = parseInt(Date.parse(end_timestamp) / 1000);
    if (isAdminUser) {
      url = `/api/task/?p=${startIdx}&channel_id=${channel_id}&task_id=${task_id}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`;
    } else {
      url = `/api/task/self?p=${startIdx}&task_id=${task_id}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`;
    }
    const res = await API.get(url);
    let { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setLogsFormat(data);
      } else {
        let newLogs = [...logs];
        newLogs.splice(startIdx * ITEMS_PER_PAGE, data.length, ...data);
        setLogsFormat(newLogs);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const pageData = logs.slice(
    (activePage - 1) * ITEMS_PER_PAGE,
    activePage * ITEMS_PER_PAGE,
  );

  const handlePageChange = (page) => {
    setActivePage(page);
    if (page === Math.ceil(logs.length / ITEMS_PER_PAGE) + 1) {
      // In this case we have to load more data and then append them.
      loadLogs(page - 1).then((r) => {});
    }
  };

  const refresh = async () => {
    // setLoading(true);
    setActivePage(1);
    await loadLogs(0);
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess('Copied: ' + text);
    } else {
      // setSearchKeyword(text);
      Modal.error({ title: t('无法复制到剪贴板，请手动复制'), content: text });
    }
  };

  useEffect(() => {
    refresh().then();
  }, [logType]);

  const renderType = (type) => {
    switch (type) {
      case 'MUSIC':
        return (
          <Label basic color='grey'>
            {' '}
            {t('生成音乐')}{' '}
          </Label>
        );
      case 'LYRICS':
        return (
          <Label basic color='pink'>
            {' '}
            {t('生成歌词')}{' '}
          </Label>
        );

      default:
        return (
          <Label basic color='black'>
            {' '}
            {t('未知')}{' '}
          </Label>
        );
    }
  };

  const renderPlatform = (type) => {
    switch (type) {
      case 'suno':
        return (
          <Label basic color='green'>
            {' '}{i18n.t('Suno')}{' '}
          </Label>
        );
      default:
        return (
          <Label basic color='black'>
            {' '}
            {t('未知')}{' '}
          </Label>
        );
    }
  };

  const renderStatus = (type) => {
    switch (type) {
      case 'SUCCESS':
        return (
          <Label basic color='green'>
            {' '}
            {t('成功')}{' '}
          </Label>
        );
      case 'NOT_START':
        return (
          <Label basic color='black'>
            {' '}
            {t('未启动')}{' '}
          </Label>
        );
      case 'SUBMITTED':
        return (
          <Label basic color='yellow'>
            {' '}
            {t('队列中')}{' '}
          </Label>
        );
      case 'IN_PROGRESS':
        return (
          <Label basic color='blue'>
            {' '}
            {t('执行中')}{' '}
          </Label>
        );
      case 'FAILURE':
        return (
          <Label basic color='red'>
            {' '}
            {t('失败')}{' '}
          </Label>
        );
      case 'QUEUED':
        return (
          <Label basic color='red'>
            {' '}
            {t('排队中')}{' '}
          </Label>
        );
      case 'UNKNOWN':
        return (
          <Label basic color='red'>
            {' '}
            {t('未知')}{' '}
          </Label>
        );
      case '':
        return (
          <Label basic color='black'>
            {' '}
            {t('正在提交')}{' '}
          </Label>
        );
      default:
        return (
          <Label basic color='black'>
            {' '}
            {t('未知')}{' '}
          </Label>
        );
    }
  };

  return (
    <>
      <Layout>
        <Form layout='horizontal' labelPosition='inset'>
          <>
            {isAdminUser && (
              <Form.Input
                field='channel_id'
                label={i18n.t('渠道 ID')}
                style={{ width: '236px', marginBottom: '10px' }}
                value={channel_id}
                placeholder={t('可选值')}
                name='channel_id'
                onChange={(value) => handleInputChange(value, 'channel_id')}
              />
            )}
            <Form.Input
              field='task_id'
              label={t('任务 ID')}
              style={{ width: '236px', marginBottom: '10px' }}
              value={task_id}
              placeholder={t('可选值')}
              name='task_id'
              onChange={(value) => handleInputChange(value, 'task_id')}
            />

            <Form.DatePicker
              field='start_timestamp'
              label={t('起始时间')}
              style={{ width: '236px', marginBottom: '10px' }}
              initValue={start_timestamp}
              value={start_timestamp}
              type='dateTime'
              name='start_timestamp'
              onChange={(value) => handleInputChange(value, 'start_timestamp')}
            />
            <Form.DatePicker
              field='end_timestamp'
              fluid
              label={t('结束时间')}
              style={{ width: '236px', marginBottom: '10px' }}
              initValue={end_timestamp}
              value={end_timestamp}
              type='dateTime'
              name='end_timestamp'
              onChange={(value) => handleInputChange(value, 'end_timestamp')}
            />
            <Button
              label={t('查询')}
              type='primary'
              htmlType='submit'
              className='btn-margin-right'
              onClick={refresh}
            >
              {t('查询')}
            </Button>
          </>
        </Form>
        <Card>
          <Table
            columns={columns}
            dataSource={pageData}
            pagination={{
              currentPage: activePage,
              pageSize: ITEMS_PER_PAGE,
              total: logCount,
              pageSizeOpts: [10, 20, 50, 100],
              onPageChange: handlePageChange,
            }}
            loading={loading}
          />
        </Card>
        <Modal
          visible={isModalOpen}
          onOk={() => setIsModalOpen(false)}
          onCancel={() => setIsModalOpen(false)}
          closable={null}
          bodyStyle={{ height: '400px', overflow: 'auto' }} // Đặt kiểu cho vùng nội dung của modal
          width={800} // Đặt chiều rộng modal
        >
          <p style={{ whiteSpace: 'pre-line' }}>{modalContent}</p>
        </Modal>
      </Layout>
    </>
  );
};

export default LogsTable;
