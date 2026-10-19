// Copyright (c) 2025 Tethys Plex
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
package relay

import (
	commonconstant "rexi-hub-v2/constant"
	"rexi-hub-v2/relay/channel"
	"rexi-hub-v2/relay/channel/ali"
	"rexi-hub-v2/relay/channel/aws"
	"rexi-hub-v2/relay/channel/baidu"
	"rexi-hub-v2/relay/channel/baidu_v2"
	"rexi-hub-v2/relay/channel/claude"
	"rexi-hub-v2/relay/channel/cloudflare"
	"rexi-hub-v2/relay/channel/cohere"
	"rexi-hub-v2/relay/channel/deepseek"
	"rexi-hub-v2/relay/channel/dify"
	"rexi-hub-v2/relay/channel/gemini"
	"rexi-hub-v2/relay/channel/jina"
	"rexi-hub-v2/relay/channel/mistral"
	"rexi-hub-v2/relay/channel/mokaai"
	"rexi-hub-v2/relay/channel/ollama"
	"rexi-hub-v2/relay/channel/openai"
	"rexi-hub-v2/relay/channel/palm"
	"rexi-hub-v2/relay/channel/perplexity"
	"rexi-hub-v2/relay/channel/siliconflow"
	"rexi-hub-v2/relay/channel/task/suno"
	"rexi-hub-v2/relay/channel/tencent"
	"rexi-hub-v2/relay/channel/vertex"
	"rexi-hub-v2/relay/channel/volcengine"
	"rexi-hub-v2/relay/channel/xai"
	"rexi-hub-v2/relay/channel/xunfei"
	"rexi-hub-v2/relay/channel/zhipu"
	"rexi-hub-v2/relay/constant"
)

func GetAdaptor(apiType int) channel.Adaptor {
	switch apiType {
	case constant.APITypeAli:
		return &ali.Adaptor{}
	case constant.APITypeAnthropic:
		return &claude.Adaptor{}
	case constant.APITypeBaidu:
		return &baidu.Adaptor{}
	case constant.APITypeGemini:
		return &gemini.Adaptor{}
	case constant.APITypeOpenAI:
		return &openai.Adaptor{}
	case constant.APITypePaLM:
		return &palm.Adaptor{}
	case constant.APITypeTencent:
		return &tencent.Adaptor{}
	case constant.APITypeXunfei:
		return &xunfei.Adaptor{}
	case constant.APITypeZhipu:
		return &zhipu.Adaptor{}
	case constant.APITypeOllama:
		return &ollama.Adaptor{}
	case constant.APITypePerplexity:
		return &perplexity.Adaptor{}
	case constant.APITypeAws:
		return &aws.Adaptor{}
	case constant.APITypeCohere:
		return &cohere.Adaptor{}
	case constant.APITypeDify:
		return &dify.Adaptor{}
	case constant.APITypeJina:
		return &jina.Adaptor{}
	case constant.APITypeCloudflare:
		return &cloudflare.Adaptor{}
	case constant.APITypeSiliconFlow:
		return &siliconflow.Adaptor{}
	case constant.APITypeVertexAi:
		return &vertex.Adaptor{}
	case constant.APITypeMistral:
		return &mistral.Adaptor{}
	case constant.APITypeDeepSeek:
		return &deepseek.Adaptor{}
	case constant.APITypeMokaAI:
		return &mokaai.Adaptor{}
	case constant.APITypeVolcEngine:
		return &volcengine.Adaptor{}
	case constant.APITypeBaiduV2:
		return &baidu_v2.Adaptor{}
	case constant.APITypeOpenRouter:
		return &openai.Adaptor{}
	case constant.APITypeXinference:
		return &openai.Adaptor{}
	case constant.APITypeGitHub:
		return &openai.Adaptor{}
	case constant.APITypeXai:
		return &xai.Adaptor{}
	}
	return nil
}

func GetTaskAdaptor(platform commonconstant.TaskPlatform) channel.TaskAdaptor {
	switch platform {
	//case constant.APITypeAIProxyLibrary:
	//	return &aiproxy.Adaptor{}
	case commonconstant.TaskPlatformSuno:
		return &suno.TaskAdaptor{}
	}
	return nil
}
