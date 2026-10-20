// Copyright (c) 2025 REXI-hub team
//
// This file is part of REXI-hub v2.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package tencent

type TencentMessage struct {
	Role    string `json:"Role"`
	Content string `json:"Content"`
}

type TencentChatRequest struct {
	// Model name, optional values include hunyuan-lite, hunyuan-standard, hunyuan-standard-256K, hunyuan-pro.
	// Để biết giới thiệu về mô hình, đọc tài liệu [Tổng quan sản phẩm](https://cloud.tencent.com/document/product/1729/104753).
	//
	// Note: 
	// Different models have different pricing, see [purchasing guide](https://cloud.tencent.com/document/product/1729/97731) and call as needed.
	Model *string `json:"Model"`
	// Chat context info。
	// Giải thích:
	// 1. Độ dài tối đa 40, sắp xếp trong mảng theo thời gian hội thoại từ cũ đến mới.
	// 2. Message.Role có các giá trị: system, user, assistant.
	// Trong đó, vai trò system là tùy chọn, nếu có thì phải nằm ở đầu danh sách. user và assistant phải xuất hiện xen kẽ (một hỏi một đáp), bắt đầu và kết thúc bằng câu hỏi của user, và Content không được rỗng. Ví dụ thứ tự Role: [system (tùy chọn) user assistant user assistant user ...].
	// 3. Tổng độ dài Content trong Messages không được vượt giới hạn độ dài đầu vào của mô hình (tham khảo tài liệu [Tổng quan sản phẩm](https://cloud.tencent.com/document/product/1729/104753)), nếu vượt sẽ cắt phần đầu và chỉ giữ phần cuối.
	Messages []*TencentMessage `json:"Messages"`
	// Streaming switch。
	// Giải thích:
	// 1. Khi không truyền giá trị thì mặc định là gọi không streaming (false).
	// 2. Khi gọi streaming thì trả kết quả tăng dần theo giao thức SSE (giá trị trả về lấy từ Choices[n].Delta, cần ghép dữ liệu tăng dần mới có kết quả đầy đủ).
	// 3. Khi gọi không streaming:
	// Cách gọi không khác gì một yêu cầu HTTP thông thường.
	// API phản hồi khá lâu, **nếu cần độ trễ thấp hơn thì nên đặt là true**.
	// Chỉ trả về kết quả cuối cùng một lần (giá trị trả về lấy từ Choices[n].Message).
	//
	// Note: 
	// Khi gọi qua SDK, gọi streaming và không streaming cần **cách khác nhau** để lấy giá trị trả về, tham khảo Note hoặc ví dụ trong SDK (ở thư mục examples/hunyuan/v20230901/ trong repo SDK của từng ngôn ngữ).
	Stream *bool `json:"Stream,omitempty"`
	// Giải thích:
	// 1. Ảnh hưởng đến độ đa dạng của văn bản đầu ra, giá trị càng lớn thì văn bản sinh ra càng đa dạng.
	// 2. Khoảng giá trị [0.0, 1.0], khi không truyền thì dùng giá trị khuyến nghị của từng mô hình.
	// 3. Không cần thiết thì không nên dùng, giá trị không hợp lý sẽ ảnh hưởng đến kết quả.
	TopP *float64 `json:"TopP,omitempty"`
	// Giải thích:
	// 1. Giá trị cao làm đầu ra ngẫu nhiên hơn, giá trị thấp làm đầu ra tập trung và xác định hơn.
	// 2. Khoảng giá trị [0.0, 2.0], khi không truyền thì dùng giá trị khuyến nghị của từng mô hình.
	// 3. Không cần thiết thì không nên dùng, giá trị không hợp lý sẽ ảnh hưởng đến kết quả.
	Temperature *float64 `json:"Temperature,omitempty"`
}

type TencentError struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
}

type TencentUsage struct {
	PromptTokens     int `json:"PromptTokens"`
	CompletionTokens int `json:"CompletionTokens"`
	TotalTokens      int `json:"TotalTokens"`
}

type TencentResponseChoices struct {
	FinishReason string         `json:"FinishReason,omitempty"` // Stream end flag; stop indicates the final packet
	Messages     TencentMessage `json:"Message,omitempty"`      // Content, chế độ đồng bộ trả về content, chế độ streaming là null. Tổng số nội dung hỗ trợ tối đa 1024 token.
	Delta        TencentMessage `json:"Delta,omitempty"`        // Content, chế độ streaming trả về content, chế độ đồng bộ là null. Tổng số nội dung hỗ trợ tối đa 1024 token.
}

type TencentChatResponse struct {
	Choices []TencentResponseChoices `json:"Choices,omitempty"` // Result
	Created int64                    `json:"Created,omitempty"` // Unix timestamp string
	Id      string                   `json:"Id,omitempty"`      // Session id
	Usage   TencentUsage             `json:"Usage,omitempty"`   // token count
	Error   TencentError             `json:"Error,omitempty"`   // Error message. Note: this field may return null, meaning no valid value
	Note    string                   `json:"Note,omitempty"`    // Note
	ReqID   string                   `json:"Req_id,omitempty"`  // Unique request Id, returned with every request. Used to trace API input
}

type TencentChatResponseSB struct {
	Response TencentChatResponse `json:"Response,omitempty"`
}
