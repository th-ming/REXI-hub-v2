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
import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Input,
  Typography,
  TextArea,
} from '@douyinfe/semi-ui';
import {
  IconPlusCircle,
  IconMinusCircle,
} from '@douyinfe/semi-icons';

const MODEL_MAPPING_EXAMPLE = {
  'gpt-3.5-turbo': 'gpt-3.5-turbo-0125',
};

// Unsafe keys to prevent prototype pollution
const UNSAFE_KEYS = new Set(['__proto__', 'prototype', 'constructor']);

/**
 * Component ModelMappingEditor - dùng để chỉnh sửa trực quan cấu hình ánh xạ mô hình
 * 
 * @param {Object} props
 * @param {string} props.value - Giá trị chuỗi JSON hiện tại
 * @param {Function} props.onChange - Hàm callback khi giá trị thay đổi
 * @param {string} props.placeholder - Văn bản placeholder
 * @param {Function} props.onRealtimeChange - Callback thay đổi thời gian thực (tùy chọn)
 */
const ModelMappingEditor = ({ value, onChange, placeholder, onRealtimeChange }) => {
  const { t } = useTranslation();
  const [mappingPairs, setMappingPairs] = useState([]);
  const [mode, setMode] = useState('visual'); // 'visual' or 'json'
  const [jsonValue, setJsonValue] = useState('');
  const [jsonError, setJsonError] = useState('');

  // Dùng để đánh dấu có phải cập nhật nội bộ hay không, tránh cập nhật vòng lặp
  const isInternalUpdateRef = useRef(false);

  // Parse JSON value to key-value pairs
  const parseJsonToMappings = (jsonStr) => {
    if (!jsonStr || jsonStr.trim() === '') {
      return [];
    }
    try {
      const parsed = JSON.parse(jsonStr);
      if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
        const unsafe = UNSAFE_KEYS;
        return Object.entries(parsed)
          .map(([key, value]) => [String(key).trim(), value])
          .filter(([k]) => k !== '' && !unsafe.has(k))
          .map(([key, value]) => ({
            id: Date.now() + Math.random(),
            key,
            value: value == null ? '' : String(value).trim(),
          }));
      }
      return [];
    } catch {
      return [];
    }
  };

  // Convert key-value pairs to JSON string
  const mappingsToJson = (pairs) => {
    if (!pairs || pairs.length === 0) {
      return '';
    }
    const obj = Object.create(null);
    const unsafe = UNSAFE_KEYS;
    pairs.forEach(pair => {
      if (pair.key && pair.key.trim() !== '') {
        const k = String(pair.key).trim();
        if (!unsafe.has(k)) {
          const v = pair.value == null ? '' : String(pair.value).trim();
          obj[k] = v;
        }
       }
    });
    return Object.keys(obj).length > 0 ? JSON.stringify(obj, null, 2) : '';
  };

  // Initialize component state from value prop
  useEffect(() => {
    // Chỉ đồng bộ trạng thái bên ngoài khi không phải cập nhật nội bộ
    if (!isInternalUpdateRef.current) {
      const pairs = parseJsonToMappings(value);
      setMappingPairs(pairs.length > 0 ? pairs : [{ id: Date.now() + Math.random(), key: '', value: '' }]);
      setJsonValue(value || '');
      setJsonError('');
    }
    isInternalUpdateRef.current = false;
  }, [value]);

  // Add new mapping pair
  const addMappingPair = () => {
    const newPairs = [...mappingPairs, { id: Date.now() + Math.random(), key: '', value: '' }];
    setMappingPairs(newPairs);
  };

  // Remove mapping pair
  const removeMappingPair = (index) => {
    const newPairs = mappingPairs.filter((_, i) => i !== index);
    const finalPairs = newPairs.length > 0 ? newPairs : [{ id: Date.now() + Math.random(), key: '', value: '' }];
    setMappingPairs(finalPairs);

    // Cập nhật component cha ngay
    const jsonStr = mappingsToJson(finalPairs);
    isInternalUpdateRef.current = true;
    onChange(jsonStr);

    // Kích hoạt đồng bộ thời gian thực
    if (onRealtimeChange) {
      onRealtimeChange(jsonStr);
    }
  };

  // Update mapping pair - chỉ cập nhật trạng thái cục bộ, không kích hoạt cập nhật component cha
  const updateMappingPair = (index, field, value) => {
    if (index < 0 || index >= mappingPairs.length) return;
    if (field !== 'key' && field !== 'value') return;
    const newPairs = [...mappingPairs];
    newPairs[index] = { ...newPairs[index], [field]: value };
    setMappingPairs(newPairs);

    // Kích hoạt callback đồng bộ thời gian thực, nhưng không cập nhật trạng thái model_mapping của component cha
    if (onRealtimeChange) {
      const jsonStr = mappingsToJson(newPairs);
      onRealtimeChange(jsonStr);
    }
  };

  // Đồng bộ lên component cha khi mất focus
  const handleInputBlur = () => {
    const jsonStr = mappingsToJson(mappingPairs);
    isInternalUpdateRef.current = true;
    onChange(jsonStr);
  };

  // Handle mode switch
  const switchMode = (newMode) => {
    if (newMode === 'json' && mode === 'visual') {
      // Switching from visual to JSON
      const jsonStr = mappingsToJson(mappingPairs);
      setJsonValue(jsonStr);
      setJsonError('');
    } else if (newMode === 'visual' && mode === 'json') {
      // Switching from JSON to visual
      if (jsonValue.trim() === '') {
        setMappingPairs([{ id: Date.now() + Math.random(), key: '', value: '' }]);
        setJsonError('');
      } else {
        const pairs = parseJsonToMappings(jsonValue);
        setMappingPairs(pairs.length > 0 ? pairs : [{ id: Date.now() + Math.random(), key: '', value: '' }]);
        setJsonError('');
      }
    }
    setMode(newMode);
  };

  // Handle JSON input change
  const handleJsonChange = (newValue) => {
    setJsonValue(newValue);

    // Validate JSON
    if (newValue.trim() === '') {
      setJsonError('');
      // Kích hoạt đồng bộ thời gian thực
      if (onRealtimeChange) {
        onRealtimeChange('');
      }
      return;
    }

    try {
      const parsed = JSON.parse(newValue);
      if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
        // Filter out unsafe keys before calling onRealtimeChange
        const unsafe = UNSAFE_KEYS;
        const filtered = Object.fromEntries(
          Object.entries(parsed).filter(([k]) => !unsafe.has(k))
        );
        setJsonError('');
        // Kích hoạt đồng bộ thời gian thực
        if (onRealtimeChange) {
          onRealtimeChange(JSON.stringify(filtered, null, 2));
        }
      } else {
        setJsonError(t('请输入有效的JSON对象格式'));
      }
    } catch (error) {
      setJsonError(t('JSON格式错误: {{message}}', { message: error.message }));
    }
  };

  // Đồng bộ lên component cha khi chế độ JSON mất focus
  const handleJsonBlur = () => {
    if (!jsonError && jsonValue.trim() !== '') {
      try {
        const parsed = JSON.parse(jsonValue);
        if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
          const sanitized = mappingsToJson(parseJsonToMappings(jsonValue));
          isInternalUpdateRef.current = true;
          onChange(sanitized);
          setJsonValue(sanitized);
        }
      } catch (error) {
        // Bỏ qua lỗi, không cập nhật component cha
      }
    } else if (jsonValue.trim() === '') {
      isInternalUpdateRef.current = true;
      onChange('');
    }
  };

  // Fill template
  const fillTemplate = () => {
    const templateJson = JSON.stringify(MODEL_MAPPING_EXAMPLE, null, 2);
    if (mode === 'visual') {
      const pairs = parseJsonToMappings(templateJson);
      setMappingPairs(pairs);
    } else {
      setJsonValue(templateJson);
    }

    // Cập nhật component cha ngay và kích hoạt đồng bộ thời gian thực
    isInternalUpdateRef.current = true;
    onChange(templateJson);
    if (onRealtimeChange) {
      onRealtimeChange(templateJson);
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 10 }}>
        <div style={{ display: 'flex', marginRight: 16 }}>
          <Button
            type={mode === 'visual' ? 'primary' : 'tertiary'}
            onClick={() => switchMode('visual')}
            style={{
              borderRadius: '6px 0 0 6px'
            }}
          >
            {t('可视化编辑')}
          </Button>
          <Button
            type={mode === 'json' ? 'primary' : 'tertiary'}
            onClick={() => switchMode('json')}
            style={{
              borderRadius: '0 6px 6px 0'
            }}
          >
            {t('JSON编辑')}
          </Button>
        </div>
        <Typography.Text
          style={{
            color: 'rgba(var(--semi-blue-5), 1)',
            userSelect: 'none',
            cursor: 'pointer',
          }}
          onClick={fillTemplate}
        >
          {t('填入模板')}
        </Typography.Text>
      </div>

      {mode === 'visual' ? (
        <div>
          <div style={{ marginBottom: 10 }}>
            <Typography.Text type="secondary">{placeholder}</Typography.Text>
          </div>

          {mappingPairs.map((pair, index) => (
            <div key={pair.id} style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
              <Input
                placeholder={t('目标模型名称')}
                value={pair.key}
                onChange={(value) => updateMappingPair(index, 'key', value)}
                onBlur={handleInputBlur}
                style={{ flex: 1, marginRight: 8 }}
              />
              <Typography.Text style={{ margin: '0 8px' }}>→</Typography.Text>
              <Input
                placeholder={t('实际模型名称')}
                value={pair.value}
                onChange={(value) => updateMappingPair(index, 'value', value)}
                onBlur={handleInputBlur}
                style={{ flex: 1, marginRight: 8 }}
              />
              <Button
                type="danger"
                icon={<IconMinusCircle />}
                size="small"
                onClick={() => removeMappingPair(index)}
                style={{ marginLeft: 4 }}
              />
            </div>
          ))}

          <Button
            type="tertiary"
            icon={<IconPlusCircle />}
            onClick={addMappingPair}
            style={{ width: '100%', marginTop: 8 }}
          >
            {t('添加映射')}
          </Button>
        </div>
      ) : (
        <div>
          <TextArea
            placeholder={
              t(
                '此项可选，用于修改请求体中的模型名称，为一个 JSON 字符串，键为请求中模型名称，值为要替换的模型名称，例如：',
              ) + `\n${JSON.stringify(MODEL_MAPPING_EXAMPLE, null, 2)}`
            }
            value={jsonValue}
            onChange={handleJsonChange}
            onBlur={handleJsonBlur}
            autosize
            autoComplete='new-password'
          />
          {jsonError && (
            <Typography.Text type="danger" style={{ marginTop: 4, display: 'block' }}>
              {jsonError}
            </Typography.Text>
          )}
        </div>
      )}
    </div>
  );
};

export default ModelMappingEditor;