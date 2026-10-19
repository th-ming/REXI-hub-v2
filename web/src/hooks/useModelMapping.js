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

import { useState, useCallback } from 'react';
import { ModelMappingUtils } from '../utils/modelMappingUtils';

/**
 * Hook quản lý ánh xạ mô hình
 * Cung cấp quản lý trạng thái và phương thức thao tác liên quan đến ánh xạ mô hình
 */
export const useModelMapping = (initialMapping = {}) => {
  // Dùng để theo dõi quan hệ ánh xạ tên gốc của mô hình { displayName: originalName }
  const [modelOriginalMapping, setModelOriginalMapping] = useState(initialMapping);

  /**
   * Cập nhật danh sách mô hình và quan hệ ánh xạ
   * @param {Array} newModels - Danh sách mô hình mới
   * @param {Object} newMapping - Quan hệ ánh xạ mới
   * @param {Function} updateModelsCallback - Hàm callback cập nhật danh sách mô hình
   */
  const updateModelsList = useCallback((newModels, newMapping, updateModelsCallback) => {
    const uniqueModels = ModelMappingUtils.filterAndDedupeModels(newModels);
    
    if (updateModelsCallback) {
      updateModelsCallback(uniqueModels);
    }
    
    setModelOriginalMapping(newMapping);
  }, []);

  /**
   * Đồng bộ ánh xạ mô hình vào cấu hình mô hình
   * @param {string} mappingValue - Cấu hình ánh xạ dạng JSON
   * @param {Array} currentModels - Danh sách mô hình hiện tại
   * @param {Function} updateModelsCallback - Hàm callback cập nhật danh sách mô hình
   */
  const syncModelMappingToModels = useCallback((mappingValue, currentModels, updateModelsCallback) => {
    const mapping = ModelMappingUtils.parseModelMapping(mappingValue);

    if (!mapping || Object.keys(mapping).length === 0) {
      // Khi ánh xạ rỗng thì khôi phục về tên gốc
      if (!currentModels || currentModels.length === 0) {
        return;
      }

      // Không có quan hệ ánh xạ thì không cần khôi phục
      if (!modelOriginalMapping || Object.keys(modelOriginalMapping).length === 0) {
        updateModelsList(currentModels, {}, updateModelsCallback);
        return;
      }

      // Khôi phục mô hình về tên gốc
      const restoredModels = ModelMappingUtils.restoreModelsToOriginal(
        currentModels, 
        modelOriginalMapping
      );

      // Xóa quan hệ ánh xạ gốc vì đã khôi phục xong
      updateModelsList(restoredModels, {}, updateModelsCallback);
      return;
    }

    // Trước khi áp dụng ánh xạ mới, khôi phục mô hình hiện tại về tên gốc
    const restoredModels = ModelMappingUtils.restoreModelsToOriginal(
      currentModels, 
      modelOriginalMapping
    );

    const { updatedModels, newMapping, hasChanges } = ModelMappingUtils.applyModelMapping(
      mapping,
      restoredModels,
      {}
    );

    if (hasChanges || !ModelMappingUtils.areModelsEqual(currentModels, updatedModels)) {
      updateModelsList(updatedModels, newMapping, updateModelsCallback);
    }
  }, [modelOriginalMapping, updateModelsList]);

  /**
   * Khởi tạo quan hệ ánh xạ mô hình
   * @param {string} mappingValue - Cấu hình ánh xạ dạng JSON
   * @param {Array} models - Danh sách mô hình
   */
  const initializeModelMapping = useCallback((mappingValue, models) => {
    const mapping = ModelMappingUtils.parseModelMapping(mappingValue);
    
    if (mapping && Object.keys(mapping).length > 0) {
    const initialMapping = Object.create(null);
    const DANGEROUS_KEYS = new Set(['__proto__', 'prototype', 'constructor']);  
      
      // Dựa trên ánh xạ mô hình và danh sách mô hình hiện tại để thiết lập quan hệ ánh xạ gốc
      Object.entries(mapping).forEach(([displayName, originalName]) => {
        const displayNameTrimmed = displayName.trim();
        const originalNameTrimmed = typeof originalName === 'string' ? originalName.trim() : null;
        
        if (displayNameTrimmed && originalNameTrimmed && !DANGEROUS_KEYS.has(displayNameTrimmed) && models.includes(displayNameTrimmed)) {
          initialMapping[displayNameTrimmed] = originalNameTrimmed;
        }
      });
      
      setModelOriginalMapping(initialMapping);
    } else {
      setModelOriginalMapping({});
    }
  }, []);

  /**
   * Đặt lại quan hệ ánh xạ
   */
  const resetModelMapping = useCallback(() => {
    setModelOriginalMapping({});
  }, []);

  return {
    modelOriginalMapping,
    updateModelsList,
    syncModelMappingToModels,
    initializeModelMapping,
    resetModelMapping,
    setModelOriginalMapping
  };
};