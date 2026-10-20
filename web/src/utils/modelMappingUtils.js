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

/**
 * Lớp tiện ích ánh xạ mô hình
 * Dùng để xử lý logic ánh xạ và chuyển đổi tên mô hình
 */
export class ModelMappingUtils {
  /**
   * Phân tích cấu hình ánh xạ mô hình
   * @param {string} mappingValue - Chuỗi cấu hình ánh xạ dạng JSON
   * @returns {Object|null} Đối tượng ánh xạ sau khi phân tích, phân tích thất bại trả về null
   */
  static parseModelMapping(mappingValue) {
    if (!mappingValue || typeof mappingValue !== 'string' || mappingValue.trim() === '') {
      return null;
    }

    try {
      const mapping = JSON.parse(mappingValue);
      if (typeof mapping !== 'object' || mapping === null) {
        return null;
      }
      return mapping;
    } catch (error) {
      console.warn('Failed to parse model redirection JSON:', error);
      return null;
    }
  }

  /**
   * Logic cốt lõi để áp dụng ánh xạ mô hình
   * @param {Object} mapping - Đối tượng cấu hình ánh xạ
   * @param {Array} currentModels - Danh sách mô hình hiện tại
   * @returns {Object} Đối tượng chứa danh sách mô hình đã cập nhật, quan hệ ánh xạ mới và có thay đổi hay không
   */
  static applyModelMapping(mapping, currentModels) {
    if (!mapping || typeof mapping !== 'object') {
      return { updatedModels: currentModels, newMapping: {}, hasChanges: false };
    }

    const updatedModels = new Set(currentModels);
    const newMapping = Object.create(null);
    let hasChanges = false;

    // Xử lý quan hệ ánh xạ mới
    Object.entries(mapping).forEach(([displayName, originalName]) => {
      const displayNameTrimmed = displayName.trim();
      const originalNameTrimmed = originalName?.trim();

      if (displayNameTrimmed && originalNameTrimmed) {
        // Thêm tên hiển thị vào danh sách mô hình (nếu chưa có)
        if (!updatedModels.has(displayNameTrimmed)) {
          updatedModels.add(displayNameTrimmed);
          hasChanges = true;
        }

        // Bỏ tên gốc, thay bằng tên hiển thị
        if (originalNameTrimmed !== displayNameTrimmed) {
          if (updatedModels.delete(originalNameTrimmed)) {
            hasChanges = true;
          }
        }

        // Thiết lập quan hệ ánh xạ: displayName -> originalName
        newMapping[displayNameTrimmed] = originalNameTrimmed;
      }
    });

    return {
      updatedModels: Array.from(updatedModels),
      newMapping,
      hasChanges
    };
  }

  /**
   * Khôi phục tên mô hình về tên gốc
   * @param {Array} currentModels - Danh sách mô hình hiện tại
   * @param {Object} originalMapping - Quan hệ ánh xạ gốc
   * @returns {Array} Danh sách mô hình sau khi khôi phục
   */
  static restoreModelsToOriginal(currentModels, originalMapping) {
    if (!currentModels || currentModels.length === 0) {
      return [];
    }

    if (!originalMapping || Object.keys(originalMapping).length === 0) {
      return currentModels;
    }

    // Dùng đối tượng không có prototype và truy cập thuộc tính an toàn, tránh ảnh hưởng của ô nhiễm prototype
    const safeMapping = Object.assign(Object.create(null), originalMapping);
    const restoredModels = currentModels.map((model) => {
      const key = typeof model === 'string' ? model.trim() : model;
      return Object.prototype.hasOwnProperty.call(safeMapping, key)
        ? safeMapping[key]
        : key;
    });

    // Loại trùng
    return Array.from(new Set(restoredModels));
  }

  /**
   * Kiểm tra hai danh sách mô hình có giống nhau không
   * @param {Array} models1 - Danh sách mô hình thứ nhất
   * @param {Array} models2 - Danh sách mô hình thứ hai
   * @returns {boolean} Có giống nhau hay không
   */
  static areModelsEqual(models1, models2) {
    if (!Array.isArray(models1) || !Array.isArray(models2)) {
      return false;
    }

    if (models1.length !== models2.length) {
      return false;
    }

    const sorted1 = [...models1].sort();
    const sorted2 = [...models2].sort();

    return JSON.stringify(sorted1) === JSON.stringify(sorted2);
  }

  /**
   * Lọc và loại bỏ trùng lặp trong danh sách mô hình
   * @param {Array} models - Danh sách mô hình
   * @returns {Array} Danh sách mô hình sau khi lọc
   */
  static filterAndDedupeModels(models) {
    if (!Array.isArray(models)) {
      return [];
    }

    return Array.from(new Set(models.filter((m) => typeof m === 'string').map((m) => m.trim()).filter(Boolean)));
  }
}