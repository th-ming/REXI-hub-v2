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
import i18n from '../../../i18n/i18n';
import React, { useEffect, useState, useRef } from 'react';
import {
  Button,
  Table,
  Modal,
  Form,
  Input,
  InputNumber,
  Space,
  Popconfirm,
  Typography,
  Banner,
  Card,
  Spin,
  Tag,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess, showWarning } from '../../../helpers';
import { useTranslation } from 'react-i18next';
import {
  IconPlus,
  IconEdit,
  IconDelete,
  IconRefresh,
} from '@douyinfe/semi-icons';

const { Title, Text } = Typography;

export default function SettingModelMapping() {
  const { t } = useTranslation();
  
  const [loading, setLoading] = useState(false);
  const [mappings, setMappings] = useState({});
  const [modalVisible, setModalVisible] = useState(false);
  const [editingMapping, setEditingMapping] = useState(null);
  const [models, setModels] = useState([{ model: '', priorities: 0 }]);
  const [virtualModelError, setVirtualModelError] = useState('');
  const [modelErrors, setModelErrors] = useState([]);
  const formApiRef = useRef();

  // Lấy cấu hình ánh xạ mô hình
  const fetchMappings = async () => {
    try {
      setLoading(true);
      const res = await API.get('/api/model_mapping/');
      if (res.data.success) {
        setMappings(res.data.data.mapping || {});
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('Failed to get model mapping config: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMappings();
  }, []);

  // Tự động kiểm tra khi models thay đổi
  useEffect(() => {
    validateModels();
  }, [models]);

  // Kiểm tra tên mô hình ảo theo thời gian thực
  const validateVirtualModel = (value) => {
    const trimmedValue = value.trim();
    if (!trimmedValue) {
      setVirtualModelError('');
      return;
    }
    
    if (!editingMapping && mappings[trimmedValue]) {
      setVirtualModelError(`Virtual model name '${trimmedValue}' already exists`);
    } else if (editingMapping && editingMapping !== trimmedValue && mappings[trimmedValue]) {
      setVirtualModelError(`Virtual model name '${trimmedValue}' already exists`);
    } else {
      setVirtualModelError('');
    }
  };

  // Kiểm tra tên mô hình thực theo thời gian thực
  const validateModels = () => {
    const errors = [];
    const modelNames = new Set();
    const duplicates = new Set();
    
    models.forEach((model, index) => {
      const trimmedModel = model.model.trim();
      if (trimmedModel) {
        if (modelNames.has(trimmedModel)) {
          duplicates.add(trimmedModel);
          errors[index] = `Duplicate model name: ${trimmedModel}`;
        } else {
          modelNames.add(trimmedModel);
          errors[index] = '';
        }
      } else {
        errors[index] = '';
      }
    });
    
    // Đặt thông báo lỗi cho các mô hình bị trùng
    models.forEach((model, index) => {
      const trimmedModel = model.model.trim();
      if (duplicates.has(trimmedModel) && !errors[index]) {
        errors[index] = `Duplicate model name: ${trimmedModel}`;
      }
    });
    
    setModelErrors(errors);
  };

  // Chuyển dữ liệu thành định dạng bảng
  const getTableData = () => {
    return Object.entries(mappings).map(([virtualModel, items]) => ({
      key: virtualModel,
      virtualModel,
      models: items || [],
    }));
  };

  // Thêm/chỉnh sửa ánh xạ
  const handleSaveMapping = async () => {
    try {
      const values = await formApiRef.current.validate();
      const { virtualModel } = values;

      // Xử lý dữ liệu danh sách mô hình
      const processedModels = models
        .map((model) => ({
          model: model.model.trim(),
          priorities: parseInt(model.priorities) || 0,
        }))
        .filter((model) => model.model); // Lọc mô hình rỗng

      // Kiểm tra ở frontend: xem tên mô hình ảo có bị trùng không
      const trimmedVirtualModel = virtualModel.trim();
      if (!editingMapping && mappings[trimmedVirtualModel]) {
        showError(`Virtual model name '${trimmedVirtualModel}' already exists, please use a different name`);
        return;
      }
      
      // Nếu ở chế độ chỉnh sửa nhưng tên mô hình ảo đã tồn tại và không phải mục đang chỉnh sửa
      if (editingMapping && editingMapping !== trimmedVirtualModel && mappings[trimmedVirtualModel]) {
        showError(`Virtual model name '${trimmedVirtualModel}' already exists, please use a different name`);
        return;
      }

      // Kiểm tra ở frontend: xem tên mô hình thực có bị trùng trong cùng một mô hình ảo không
      const modelNames = new Set();
      const duplicateModels = [];
      
      for (let i = 0; i < processedModels.length; i++) {
        const modelName = processedModels[i].model;
        if (modelNames.has(modelName)) {
          duplicateModels.push(modelName);
        } else {
          modelNames.add(modelName);
        }
      }
      
      if (duplicateModels.length > 0) {
        showError(`Duplicate actual model names: ${duplicateModels.join(', ')}`);
        return;
      }

      // Kiểm tra có tên mô hình thực nào rỗng không
      if (processedModels.length === 0) {
        showError(i18n.t('At least one actual model is required'));
        return;
      }

      const newMappings = { ...mappings };

      // Nếu ở chế độ chỉnh sửa và tên mô hình ảo đã thay đổi thì cần xóa ánh xạ cũ
      if (editingMapping && editingMapping !== trimmedVirtualModel) {
        delete newMappings[editingMapping];
      }

      newMappings[trimmedVirtualModel] = processedModels;

      const res = await API.put('/api/model_mapping/', {
        mapping: newMappings,
      });
      if (res.data.success) {
        showSuccess(res.data.message || i18n.t('Saved successfully'));
        setMappings(newMappings);
        setModalVisible(false);
        if (formApiRef.current) {
          formApiRef.current.reset();
        }
        setEditingMapping(null);
        setModels([{ model: '', priorities: 0 }]);
        setVirtualModelError('');
        setModelErrors([]);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      if (error.errorFields) {
        showWarning('Please check the form input');
      } else {
        showError('Save failed: ' + error.message);
      }
    }
  };

  // Xóa ánh xạ
  const handleDeleteMapping = async (virtualModel) => {
    try {
      const newMappings = { ...mappings };
      delete newMappings[virtualModel];

      const res = await API.put('/api/model_mapping/', {
        mapping: newMappings,
      });
      if (res.data.success) {
        showSuccess(i18n.t('Deleted successfully'));
        setMappings(newMappings);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('Delete failed: ' + error.message);
    }
  };

  // Tải lại cấu hình
  const handleReload = async () => {
    try {
      const res = await API.post('/api/model_mapping/reload');
      if (res.data.success) {
        showSuccess(res.data.message);
        await fetchMappings();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('Reload failed: ' + error.message);
    }
  };

  // Mở modal chỉnh sửa
  const openEditModal = (virtualModel = null) => {
    setEditingMapping(virtualModel);
    setModalVisible(true);
    setVirtualModelError('');
    setModelErrors([]);

    // Đặt giá trị form muộn một chút để đảm bảo modal đã render
    setTimeout(() => {
      if (formApiRef.current) {
        if (virtualModel && mappings[virtualModel]) {
          formApiRef.current.setValues({
            virtualModel,
          });
          setModels(mappings[virtualModel] || [{ model: '', priorities: 0 }]);
        } else {
          formApiRef.current.reset();
          setModels([{ model: '', priorities: 0 }]);
        }
      }
    }, 100);
  };

  // Thêm mô hình
  const addModel = () => {
    setModels([...models, { model: '', priorities: 0 }]);
  };

  // Xóa mô hình
  const removeModel = (index) => {
    if (models.length > 1) {
      const newModels = models.filter((_, i) => i !== index);
      setModels(newModels);
    }
  };

  // Cập nhật mô hình
  const updateModel = (index, field, value) => {
    const newModels = [...models];
    newModels[index][field] = value;
    setModels(newModels);
  };

  // Sao chép tên mô hình ảo
  const copyVirtualModel = (virtualModel) => {
    navigator.clipboard
      .writeText(virtualModel)
      .then(() => {
        showSuccess(`Copied: ${virtualModel}`);
      })
      .catch(() => {
        // Phương án dự phòng
        const textarea = document.createElement('textarea');
        textarea.value = virtualModel;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
        showSuccess(`Copied: ${virtualModel}`);
      });
  };

  const columns = [
    {
      title: t('虚拟模型名'),
      dataIndex: 'virtualModel',
      key: 'virtualModel',
      width: 200,
      render: (virtualModel) => (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Tag
            color='violet'
            size='large'
            style={{
              fontSize: '13px',
              fontWeight: '500',
              padding: '4px 12px',
              borderRadius: '6px',
              cursor: 'pointer',
              userSelect: 'none',
              transition: 'all 0.2s',
            }}
            onClick={() => copyVirtualModel(virtualModel)}
            onMouseEnter={(e) => {
              e.target.style.transform = 'scale(1.02)';
              e.target.style.boxShadow = '0 2px 8px rgba(0,0,0,0.15)';
            }}
            onMouseLeave={(e) => {
              e.target.style.transform = 'scale(1)';
              e.target.style.boxShadow = 'none';
            }}
            title={i18n.t('点击复制模型名称')}
          >
            {virtualModel}
          </Tag>
        </div>
      ),
    },
    {
      title: t('实际模型映射'),
      dataIndex: 'models',
      key: 'models',
      render: (models) => {
        // Tìm giá trị có độ ưu tiên cao nhất
        const maxPriority = Math.max(...models.map(model => model.priorities));
        
        return (
          <Space wrap>
            {models.map((model, index) => {
              // Kiểm tra có phải độ ưu tiên cao nhất không
              const isHighestPriority = model.priorities === maxPriority;
              
              return (
                <Tag
                  key={index}
                  color={
                    isHighestPriority
                      ? 'green'
                      : model.priorities >= 8
                        ? 'blue'
                        : model.priorities >= 5
                          ? 'orange'
                          : 'grey'
                  }
                  style={
                    isHighestPriority
                      ? { fontWeight: 'bold' }
                      : {}
                  }
                >
                  {t('{{model}} (优先级: {{priority}})', { model: model.model, priority: model.priorities })}
                </Tag>
              );
            })}
          </Space>
        );
      },
    },
    {
      title: t('操作'),
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button
            icon={<IconEdit />}
            size='small'
            onClick={() => openEditModal(record.virtualModel)}
          >
            {t('编辑')}
          </Button>
          <Popconfirm
            title={i18n.t('Delete this model mapping?')}
            onConfirm={() => handleDeleteMapping(record.virtualModel)}
            position='leftTop'
          >
            <Button icon={<IconDelete />} type='danger' size='small'>
              {t('删除')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Spin spinning={loading}>
        <Form style={{ marginBottom: 15 }}>
          <Form.Section text={i18n.t('Model Mapping Settings')}>
            <Banner
              type='info'
              description={i18n.t('Configure the mapping from virtual model names to actual models. Supports multiple actual models selected by priority and round-robin.')}
              style={{ marginBottom: 16 }}
            />

            <Space style={{ marginBottom: 16 }}>
              <Button
                type='primary'
                icon={<IconPlus />}
                onClick={() => openEditModal()}
              >
                {t('添加映射')}
              </Button>
              <Button icon={<IconRefresh />} onClick={handleReload}>
                {t('重新加载')}
              </Button>
            </Space>

            <div style={{ position: 'relative', overflow: 'visible' }}>
              <Table
                columns={columns}
                dataSource={getTableData()}
                pagination={false}
                size='small'
                empty={
                  <div style={{ textAlign: 'center', padding: '20px' }}>
                    <Text type='secondary'>{t('暂无模型映射配置')}</Text>
                  </div>
                }
              />
            </div>
          </Form.Section>
        </Form>
      </Spin>

      {/* Modal thêm/chỉnh sửa */}
      <Modal
        title={editingMapping ? i18n.t('Edit Model Mapping') : i18n.t('Add Model Mapping')}
        visible={modalVisible}
        onOk={handleSaveMapping}
        onCancel={() => {
          setModalVisible(false);
          setEditingMapping(null);
          setModels([{ model: '', priorities: 0 }]);
          setVirtualModelError('');
          setModelErrors([]);
          if (formApiRef.current) {
            formApiRef.current.reset();
          }
        }}
        width={600}
      >
        <Form
          getFormApi={(formAPI) => (formApiRef.current = formAPI)}
          labelPosition='left'
          labelWidth={120}
        >
          <Form.Input
            field='virtualModel'
            label={i18n.t('虚拟模型名')}
            placeholder={i18n.t('Enter virtual model name, e.g. gpt-4')}
            rules={[
              { required: true, message: 'Please enter a virtual model name' },
              { type: 'string', message: i18n.t('Virtual model name must be a string') },
            ]}
            disabled={false}
            onChange={validateVirtualModel}
            validateStatus={virtualModelError ? 'error' : ''}
            helpText={virtualModelError}
          />
          <div style={{ marginBottom: 16 }}>
            <Text strong>{t('实际模型列表:')}</Text>
            {models.map((model, index) => (
              <div
                key={index}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  marginTop: 8,
                  marginBottom: 8,
                }}
              >
                <div style={{ display: 'flex', flexDirection: 'column', marginRight: 8 }}>
                  <Input
                    value={model.model}
                    placeholder={i18n.t('实际模型名称')}
                    style={{ 
                      width: 300,
                      borderColor: modelErrors[index] ? '#ff4d4f' : undefined
                    }}
                    onChange={(value) => updateModel(index, 'model', value)}
                  />
                  {modelErrors[index] && (
                    <Text 
                      type="danger" 
                      size="small" 
                      style={{ fontSize: '12px', marginTop: '2px' }}
                    >
                      {modelErrors[index]}
                    </Text>
                  )}
                </div>
                <InputNumber
                  value={model.priorities}
                  placeholder={i18n.t('优先级')}
                  min={0}
                  style={{ width: 100, marginRight: 8, alignSelf: 'flex-start' }}
                  onChange={(value) => updateModel(index, 'priorities', value)}
                />
                <Button
                  type='danger'
                  size='small'
                  onClick={() => removeModel(index)}
                  disabled={models.length === 1}
                  style={{ alignSelf: 'flex-start' }}
                >
                  {t('删除')}
                </Button>
              </div>
            ))}
            <Button
              onClick={addModel}
              icon={<IconPlus />}
              size='small'
              style={{ marginTop: 8 }}
            >
              {t('添加模型')}
            </Button>
          </div>
        </Form>
      </Modal>
    </>
  );
}
