/*
Copyright (c) 2025 REXI-hub team

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
import React, { useEffect, useState } from 'react';
import {
  Table,
  Button,
  Input,
  Modal,
  Form,
  Space,
  Typography,
  Radio,
  Notification,
} from '@douyinfe/semi-ui';
import {
  IconDelete,
  IconPlus,
  IconSearch,
  IconSave,
  IconBolt,
} from '@douyinfe/semi-icons';
import { showError, showSuccess } from '../../../helpers';
import { API } from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function ModelRatioNotSetEditor(props) {
  const { t } = useTranslation();
  const [models, setModels] = useState([]);
  const [visible, setVisible] = useState(false);
  const [batchVisible, setBatchVisible] = useState(false);
  const [currentModel, setCurrentModel] = useState(null);
  const [searchText, setSearchText] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [enabledModels, setEnabledModels] = useState([]);
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);
  const [batchFillType, setBatchFillType] = useState('ratio');
  const [batchFillValue, setBatchFillValue] = useState('');
  const [batchRatioValue, setBatchRatioValue] = useState('');
  const [batchCompletionRatioValue, setBatchCompletionRatioValue] =
    useState('');
  const { Text } = Typography;
  // Định nghĩa số mục hiển thị mỗi trang (tùy chọn)
  const pageSizeOptions = [10, 20, 50, 100];

  const getAllEnabledModels = async () => {
    try {
      const res = await API.get('/api/channel/models_enabled');
      const { success, message, data } = res.data;
      if (success) {
        setEnabledModels(data);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error(t('获取启用模型失败:'), error);
      showError(t('获取启用模型失败'));
    }
  };

  useEffect(() => {
    // Lấy toàn bộ mô hình đang bật
    getAllEnabledModels();
  }, []);

  useEffect(() => {
    try {
      const modelPrice = JSON.parse(props.options.ModelPrice || '{}');
      const modelRatio = JSON.parse(props.options.ModelRatio || '{}');
      const completionRatio = JSON.parse(props.options.CompletionRatio || '{}');

      // Tìm ra toàn bộ mô hình chưa đặt giá và hệ số
      const unsetModels = enabledModels.filter((modelName) => {
        const hasPrice = modelPrice[modelName] !== undefined;
        const hasRatio = modelRatio[modelName] !== undefined;

        // Nếu mô hình không có giá hoặc chưa đặt hệ số thì hiển thị
        return !hasPrice && !hasRatio;
      });

      // Tạo dữ liệu mô hình
      const modelData = unsetModels.map((name) => ({
        name,
        price: modelPrice[name] || '',
        ratio: modelRatio[name] || '',
        completionRatio: completionRatio[name] || '',
      }));

      setModels(modelData);
      // Xóa lựa chọn
      setSelectedRowKeys([]);
    } catch (error) {
      console.error(t('JSON解析错误:'), error);
    }
  }, [props.options, enabledModels]);

  // Trước tiên khai báo các hàm tiện ích liên quan đến phân trang
  const getPagedData = (data, currentPage, pageSize) => {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    return data.slice(start, end);
  };

  // Xử lý thay đổi kích thước trang
  const handlePageSizeChange = (size) => {
    setPageSize(size);
    // Tính lại trang hiện tại để tránh mất dữ liệu
    const totalPages = Math.ceil(filteredModels.length / size);
    if (currentPage > totalPages) {
      setCurrentPage(totalPages || 1);
    }
  };

  // Trước câu lệnh return thì xử lý logic lọc và phân trang trước
  const filteredModels = models.filter((model) =>
    searchText
      ? model.name.toLowerCase().includes(searchText.toLowerCase())
      : true,
  );

  // Sau đó tính dữ liệu phân trang dựa trên dữ liệu đã lọc
  const pagedData = getPagedData(filteredModels, currentPage, pageSize);

  const SubmitData = async () => {
    setLoading(true);
    const output = {
      ModelPrice: JSON.parse(props.options.ModelPrice || '{}'),
      ModelRatio: JSON.parse(props.options.ModelRatio || '{}'),
      CompletionRatio: JSON.parse(props.options.CompletionRatio || '{}'),
    };

    try {
      // Chuyển đổi dữ liệu - chỉ xử lý các mô hình đã sửa
      models.forEach((model) => {
        // Chỉ cập nhật khi người dùng đã đặt giá trị
        if (model.price !== '') {
          // Nếu giá không rỗng thì chuyển thành số thực, bỏ qua tham số hệ số
          output.ModelPrice[model.name] = parseFloat(model.price);
        } else {
          if (model.ratio !== '')
            output.ModelRatio[model.name] = parseFloat(model.ratio);
          if (model.completionRatio !== '')
            output.CompletionRatio[model.name] = parseFloat(
              model.completionRatio,
            );
        }
      });

      // Chuẩn bị mảng yêu cầu API
      const finalOutput = {
        ModelPrice: JSON.stringify(output.ModelPrice, null, 2),
        ModelRatio: JSON.stringify(output.ModelRatio, null, 2),
        CompletionRatio: JSON.stringify(output.CompletionRatio, null, 2),
      };

      const requestQueue = Object.entries(finalOutput).map(([key, value]) => {
        return API.put('/api/option/', {
          key,
          value,
        });
      });

      // Xử lý hàng loạt yêu cầu
      const results = await Promise.all(requestQueue);

      // Kiểm tra kết quả
      if (requestQueue.length === 1) {
        if (results.includes(undefined)) return;
      } else if (requestQueue.length > 1) {
        if (results.includes(undefined)) {
          return showError(t('部分保存失败，请重试'));
        }
      }

      // Kiểm tra kết quả của từng yêu cầu
      for (const res of results) {
        if (!res.data.success) {
          return showError(res.data.message);
        }
      }

      showSuccess(t('保存成功'));
      props.refresh();
      // Lấy lại các mô hình chưa được đặt
      getAllEnabledModels();
    } catch (error) {
      console.error(t('保存失败:'), error);
      showError(t('保存失败，请重试'));
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    {
      title: t('模型名称'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('模型固定价格'),
      dataIndex: 'price',
      key: 'price',
      render: (text, record) => (
        <Input
          value={text}
          placeholder={t('按量计费')}
          onChange={(value) => updateModel(record.name, 'price', value)}
        />
      ),
    },
    {
      title: t('模型倍率'),
      dataIndex: 'ratio',
      key: 'ratio',
      render: (text, record) => (
        <Input
          value={text}
          placeholder={record.price !== '' ? t('模型倍率') : t('输入模型倍率')}
          disabled={record.price !== ''}
          onChange={(value) => updateModel(record.name, 'ratio', value)}
        />
      ),
    },
    {
      title: t('补全倍率'),
      dataIndex: 'completionRatio',
      key: 'completionRatio',
      render: (text, record) => (
        <Input
          value={text}
          placeholder={record.price !== '' ? t('补全倍率') : t('输入补全倍率')}
          disabled={record.price !== ''}
          onChange={(value) =>
            updateModel(record.name, 'completionRatio', value)
          }
        />
      ),
    },
  ];
  // Mới: điền hệ số còn thiếu từ preset trang chính thức mixed (chỉ với mục hiện chưa đặt)
  const [loadingMixed, setLoadingMixed] = useState(false);
  const handleFillFromMixed = async () => {
    try {
      setLoadingMixed(true);
      // Tham khảo nguồn và cấu trúc của ModelCommonRatioSettings.fetchPresetRatios
      const promptUrl = 'https://public-assets.rexi-hub-v2.org/defaults/model-ratios/mixed/prompt.json';
      const completionUrl = 'https://public-assets.rexi-hub-v2.org/defaults/model-ratios/mixed/completion.json';

      const [promptRatios, completionRatios] = await Promise.all([
        fetch(promptUrl).then((res) => (res.ok ? res.json() : {})).catch(() => ({})),
        fetch(completionUrl).then((res) => (res.ok ? res.json() : {})).catch(() => ({})),
      ]);

      const bothEmpty =
        (!promptRatios || Object.keys(promptRatios).length === 0) &&
        (!completionRatios || Object.keys(completionRatios).length === 0);
      if (bothEmpty) {
        showError(t('填入官网倍率失败，请稍后重试'));
        return;
      }

      let filledCount = 0;

      setModels((prev) =>
        prev.map((m) => {
          // Chỉ xét điền khi mô hình có trong danh mục (tức có trong prev); mô hình chưa được thu thập đương nhiên không xuất hiện trong prev
          // Ưu tiên giữ price: nếu đã đặt price thì không điền hệ số
          if (m.price !== '') return m;

          const next = { ...m };
          if ((next.ratio === '' || next.ratio === undefined) && promptRatios && promptRatios[next.name] !== undefined) {
            next.ratio = String(promptRatios[next.name]);
            filledCount += 1;
          }
          if ((next.completionRatio === '' || next.completionRatio === undefined) && completionRatios && completionRatios[next.name] !== undefined) {
            next.completionRatio = String(completionRatios[next.name]);
            filledCount += 1;
          }
          return next;
        }),
      );

      if (filledCount > 0) {
        Notification.success({
          title: t('已填入官网倍率'),
          content: t('成功填充 {{count}} 项缺失倍率', { count: filledCount }),
          duration: 3,
        });
      } else {
        Notification.info({
          title: t('无可填充项'),
          content: t('未找到可从官网填充的缺失倍率或均已设置'),
          duration: 3,
        });
      }
    } catch (e) {
      console.error('Failed to fill official ratios:', e);
      showError(t('填入官网倍率失败，请稍后重试'));
    } finally {
      setLoadingMixed(false);
    }
  };

  const updateModel = (name, field, value) => {
    if (value !== '' && isNaN(value)) {
      showError(t('请输入数字'));
      return;
    }
    setModels((prev) =>
      prev.map((model) =>
        model.name === name ? { ...model, [field]: value } : model,
      ),
    );
  };

  const addModel = (values) => {
    // Kiểm tra tên mô hình đã tồn tại chưa, nếu có thì từ chối thêm
    if (models.some((model) => model.name === values.name)) {
      showError(t('模型名称已存在'));
      return;
    }
    setModels((prev) => [
      {
        name: values.name,
        price: values.price || '',
        ratio: values.ratio || '',
        completionRatio: values.completionRatio || '',
      },
      ...prev,
    ]);
    setVisible(false);
    showSuccess(t('添加成功'));
  };

  // Chức năng điền hàng loạt
  const handleBatchFill = () => {
    if (selectedRowKeys.length === 0) {
      showError(t('请先选择需要批量设置的模型'));
      return;
    }

    if (batchFillType === 'bothRatio') {
      if (batchRatioValue === '' || batchCompletionRatioValue === '') {
        showError(t('请输入模型倍率和补全倍率'));
        return;
      }
      if (isNaN(batchRatioValue) || isNaN(batchCompletionRatioValue)) {
        showError(t('请输入有效的数字'));
        return;
      }
    } else {
      if (batchFillValue === '') {
        showError(t('请输入填充值'));
        return;
      }
      if (isNaN(batchFillValue)) {
        showError(t('请输入有效的数字'));
        return;
      }
    }

    // Cập nhật hàng loạt mô hình theo loại đã chọn
    setModels((prev) =>
      prev.map((model) => {
        if (selectedRowKeys.includes(model.name)) {
          if (batchFillType === 'price') {
            return {
              ...model,
              price: batchFillValue,
              ratio: '',
              completionRatio: '',
            };
          } else if (batchFillType === 'ratio') {
            return {
              ...model,
              price: '',
              ratio: batchFillValue,
            };
          } else if (batchFillType === 'completionRatio') {
            return {
              ...model,
              price: '',
              completionRatio: batchFillValue,
            };
          } else if (batchFillType === 'bothRatio') {
            return {
              ...model,
              price: '',
              ratio: batchRatioValue,
              completionRatio: batchCompletionRatioValue,
            };
          }
        }
        return model;
      }),
    );

    setBatchVisible(false);
    Notification.success({
      title: t('批量设置成功'),
      content: t('已为 {{count}} 个模型设置{{type}}', {
        count: selectedRowKeys.length,
        type:
          batchFillType === 'price'
            ? t('固定价格')
            : batchFillType === 'ratio'
              ? t('模型倍率')
              : batchFillType === 'completionRatio'
                ? t('补全倍率')
                : t('模型倍率和补全倍率'),
      }),
      duration: 3,
    });
  };

  const handleBatchTypeChange = (value) => {
    console.log(t('Changing batch type to:'), value);
    setBatchFillType(value);

    // Khi đổi loại thì xóa trắng giá trị tương ứng
    if (value !== 'bothRatio') {
      setBatchFillValue('');
    } else {
      setBatchRatioValue('');
      setBatchCompletionRatioValue('');
    }
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: (selectedKeys) => {
      setSelectedRowKeys(selectedKeys);
    },
  };

  return (
    <>
      <Space vertical align='start' style={{ width: '100%' }}>
        <Space>
          <Button icon={<IconPlus />} onClick={() => setVisible(true)}>
            {t('添加模型')}
          </Button>
          <Button
            icon={<IconBolt />}
            type='secondary'
            onClick={() => setBatchVisible(true)}
            disabled={selectedRowKeys.length === 0}
          >
            {t('批量设置')} ({selectedRowKeys.length})
          </Button>
          {/* Mới: nút điền hệ số từ trang chính thức, chỉ điền mục còn thiếu, nguồn mixed */}
          <Button
            type='tertiary'
            onClick={handleFillFromMixed}
            loading={loadingMixed}
            disabled={models.length === 0}
          >
            {t('填入官网倍率')}
          </Button>
          <Button
            type='primary'
            icon={<IconSave />}
            onClick={SubmitData}
            loading={loading}
          >
            {t('应用更改')}
          </Button>
          <Input
            prefix={<IconSearch />}
            placeholder={t('搜索模型名称')}
            value={searchText}
            onChange={(value) => {
              setSearchText(value);
              setCurrentPage(1);
            }}
            style={{ width: 200 }}
          />
        </Space>

        <Text>
          {t('此页面仅显示未设置价格或倍率的模型，设置后将自动从列表中移除')}
        </Text>

        <Table
          columns={columns}
          dataSource={pagedData}
          rowSelection={rowSelection}
          rowKey='name'
          pagination={{
            currentPage: currentPage,
            pageSize: pageSize,
            total: filteredModels.length,
            onPageChange: (page) => setCurrentPage(page),
            onPageSizeChange: handlePageSizeChange,
            pageSizeOptions: pageSizeOptions,
            formatPageText: (page) =>
              t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
                start: page.currentStart,
                end: page.currentEnd,
                total: filteredModels.length,
              }),
            showTotal: true,
            showSizeChanger: true,
          }}
          empty={
            <div style={{ textAlign: 'center', padding: '20px' }}>
              {t('没有未设置的模型')}
            </div>
          }
        />
      </Space>

      {/* Popup thêm mô hình */}
      <Modal
        title={t('添加模型')}
        visible={visible}
        onCancel={() => setVisible(false)}
        onOk={() => {
          currentModel && addModel(currentModel);
        }}
      >
        <Form>
          <Form.Input
            field='name'
            label={t('模型名称')}
            placeholder='strawberry'
            required
            onChange={(value) =>
              setCurrentModel((prev) => ({ ...prev, name: value }))
            }
          />
          <Form.Switch
            field='priceMode'
            label={
              <>
                {t('定价模式')}：
                {currentModel?.priceMode ? t('固定价格') : t('倍率模式')}
              </>
            }
            onChange={(checked) => {
              setCurrentModel((prev) => ({
                ...prev,
                price: '',
                ratio: '',
                completionRatio: '',
                priceMode: checked,
              }));
            }}
          />
          {currentModel?.priceMode ? (
            <Form.Input
              field='price'
              label={t('固定价格(每次)')}
              placeholder={t('输入每次价格')}
              onChange={(value) =>
                setCurrentModel((prev) => ({ ...prev, price: value }))
              }
            />
          ) : (
            <>
              <Form.Input
                field='ratio'
                label={t('模型倍率')}
                placeholder={t('输入模型倍率')}
                onChange={(value) =>
                  setCurrentModel((prev) => ({ ...prev, ratio: value }))
                }
              />
              <Form.Input
                field='completionRatio'
                label={t('补全倍率')}
                placeholder={t('输入补全价格')}
                onChange={(value) =>
                  setCurrentModel((prev) => ({
                    ...prev,
                    completionRatio: value,
                  }))
                }
              />
            </>
          )}
        </Form>
      </Modal>

      {/* Popup cài đặt hàng loạt */}
      <Modal
        title={t('批量设置模型参数')}
        visible={batchVisible}
        onCancel={() => setBatchVisible(false)}
        onOk={handleBatchFill}
        width={500}
      >
        <Form>
          <Form.Section text={t('设置类型')}>
            <div style={{ marginBottom: '16px' }}>
              <Space>
                <Radio
                  checked={batchFillType === 'price'}
                  onChange={() => handleBatchTypeChange('price')}
                >
                  {t('固定价格')}
                </Radio>
                <Radio
                  checked={batchFillType === 'ratio'}
                  onChange={() => handleBatchTypeChange('ratio')}
                >
                  {t('模型倍率')}
                </Radio>
                <Radio
                  checked={batchFillType === 'completionRatio'}
                  onChange={() => handleBatchTypeChange('completionRatio')}
                >
                  {t('补全倍率')}
                </Radio>
                <Radio
                  checked={batchFillType === 'bothRatio'}
                  onChange={() => handleBatchTypeChange('bothRatio')}
                >
                  {t('模型倍率和补全倍率同时设置')}
                </Radio>
              </Space>
            </div>
          </Form.Section>

          {batchFillType === 'bothRatio' ? (
            <>
              <Form.Input
                field='batchRatioValue'
                label={t('模型倍率值')}
                placeholder={t('请输入模型倍率')}
                value={batchRatioValue}
                onChange={(value) => setBatchRatioValue(value)}
              />
              <Form.Input
                field='batchCompletionRatioValue'
                label={t('补全倍率值')}
                placeholder={t('请输入补全倍率')}
                value={batchCompletionRatioValue}
                onChange={(value) => setBatchCompletionRatioValue(value)}
              />
            </>
          ) : (
            <Form.Input
              field='batchFillValue'
              label={
                batchFillType === 'price'
                  ? t('固定价格值')
                  : batchFillType === 'ratio'
                    ? t('模型倍率值')
                    : t('补全倍率值')
              }
              placeholder={t('请输入数值')}
              value={batchFillValue}
              onChange={(value) => setBatchFillValue(value)}
            />
          )}

          <Text type='tertiary'>
            {t('将为选中的 ')} <Text strong>{selectedRowKeys.length}</Text>{' '}
            {t(' 个模型设置相同的值')}
          </Text>
          <div style={{ marginTop: '8px' }}>
            <Text type='tertiary'>
              {t('当前设置类型: ')}{' '}
              <Text strong>
                {batchFillType === 'price'
                  ? t('固定价格')
                  : batchFillType === 'ratio'
                    ? t('模型倍率')
                    : batchFillType === 'completionRatio'
                      ? t('补全倍率')
                      : t('模型倍率和补全倍率')}
              </Text>
            </Text>
          </div>
        </Form>
      </Modal>
    </>
  );
}
