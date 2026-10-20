<div align="center">

![REXI-hub v2](/web/public/logo.png)

# REXI-hub v2

**Cổng tổng hợp API AI — quản lý & phân phối khóa**

[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](./LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23-blue)](https://go.dev/)
[![React](https://img.shields.io/badge/react-18-61dafb)](https://reactjs.org/)

</div>

---

## Giới thiệu

REXI-hub v2 là cổng tổng hợp API AI hỗ trợ nhiều kênh (Azure, OpenAI, Claude, Gemini...), dành cho quản lý và phân phối khóa API. Một tệp thực thi duy nhất, image Docker dựng sẵn, triển khai một chạm.

**Đặc điểm chính:**

- Hỗ trợ nhiều kênh API (OpenAI, Azure, Claude, Gemini, Cohere, Jina...)
- Quản lý khóa API, phân quyền theo nhóm
- Bảng điều khiển thống kê chi tiết
- Đa ngôn ngữ: Tiếng Việt, 中文, English
- Giao diện tối/sáng
- Hỗ trợ Rerank, Realtime, Claude Messages, Midjourney, Suno
- Tương thích One API / New API

## Triển khai

### Docker Compose (khuyến nghị)

```shell
git clone https://github.com/th-ming/REXI-hub-v2.git
cd REXI-hub-v2
docker-compose up -d
```

### Build thủ công

```shell
# Frontend
cd web && pnpm install && pnpm build && cd ..

# Backend
go build -o main.exe .

# Chạy
./main.exe
```

Server mặc định chạy tại `http://localhost:3000`.

## Biến môi trường

| Biến | Mặc định | Mô tả |
|------|----------|-------|
| `GENERATE_DEFAULT_TOKEN` | `false` | Tạo token ban đầu cho user mới |
| `STREAMING_TIMEOUT` | `60` | Timeout streaming (giây) |
| `CRYPTO_SECRET` | — | Khóa mã hóa nội dung DB |
| `SESSION_SECRET` | — | Secret cho session (bắt buộc khi multi-instance) |
| `REDIS_CONN_STRING` | — | Chuỗi kết nối Redis (cache) |
| `MEMORY_CACHE_ENABLED` | `false` | Bật cache bộ nhớ |
| `FORCE_STREAM_OPTION` | `true` | Ghi đè stream_options của client |
| `GET_MEDIA_TOKEN` | `true` | Tính token ảnh |
| `UPDATE_TASK` | `true` | Cập nhật tác vụ async (Midjourney, Suno) |
| `AZURE_DEFAULT_API_VERSION` | `2024-12-01-preview` | API version Azure mặc định |

## Công nghệ

- **Backend:** Go + GORM (SQLite / MySQL / PostgreSQL)
- **Frontend:** React 18 + Semi Design
- **Build:** Vite + pnpm

## Giấy phép

**GPL v3** — xem [LICENSE](./LICENSE).

Dựa trên [New API](https://github.com/QuantumNous/new-api) (giấy phép gốc Apache 2.0 giữ tại [LICENSE.new-api](./LICENSE.new-api)), fork từ [Veloera](https://github.com/Veloera/Veloera).

---

<div align="center">
Made with ❤️ by REXI-hub team
</div>
