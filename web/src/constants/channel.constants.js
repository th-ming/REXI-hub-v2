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

import i18n from '../i18n/i18n';
export const CHANNEL_OPTIONS = [
  { value: 1, color: 'green', label: i18n.t('OpenAI') },
  {
    value: 2,
    color: 'light-blue',
    label: i18n.t('Midjourney Proxy'),
  },
  {
    value: 5,
    color: 'blue',
    label: i18n.t('Midjourney Proxy Plus'),
  },
  {
    value: 36,
    color: 'purple',
    label: i18n.t('Suno API'),
  },
  { value: 4, color: 'grey', label: i18n.t('Ollama') },
  {
    value: 14,
    color: 'indigo',
    label: i18n.t('Anthropic Claude'),
  },
  {
    value: 33,
    color: 'indigo',
    label: i18n.t('AWS Claude'),
  },
  { value: 41, color: 'blue', label: i18n.t('Vertex AI') },
  {
    value: 3,
    color: 'teal',
    label: i18n.t('Azure OpenAI'),
  },
  {
    value: 34,
    color: 'purple',
    label: i18n.t('Cohere'),
  },
  { value: 39, color: 'grey', label: i18n.t('Cloudflare') },
  { value: 43, color: 'blue', label: i18n.t('DeepSeek') },
  {
    value: 15,
    color: 'blue',
    label: i18n.t('Baidu Wenxin Qianfan'),
  },
  {
    value: 46,
    color: 'blue',
    label: i18n.t('Baidu Wenxin Qianfan V2'),
  },
  {
    value: 17,
    color: 'orange',
    label: i18n.t('Alibaba Qwen'),
  },
  {
    value: 18,
    color: 'blue',
    label: i18n.t('iFlytek Spark'),
  },
  {
    value: 16,
    color: 'purple',
    label: i18n.t('Zhipu BigModel'),
  },
  {
    value: 24,
    color: 'orange',
    label: i18n.t('Google Gemini'),
  },
  {
    value: 11,
    color: 'orange',
    label: i18n.t('Google PaLM2'),
  },
  {
    value: 47,
    color: 'blue',
    label: i18n.t('Xinference'),
  },
  { value: 25, color: 'green', label: i18n.t('Moonshot') },
  { value: 20, color: 'green', label: i18n.t('OpenRouter') },
  { value: 49, color: 'green', label: i18n.t('GitHub Models') },
  { value: 19, color: 'blue', label: i18n.t('360 Brain') },
  { value: 23, color: 'teal', label: i18n.t('Tencent Hunyuan') },
  { value: 31, color: 'green', label: i18n.t('01.AI (Yi)') },
  { value: 35, color: 'green', label: i18n.t('MiniMax') },
  { value: 37, color: 'teal', label: i18n.t('Dify') },
  { value: 38, color: 'blue', label: i18n.t('Jina') },
  { value: 40, color: 'purple', label: i18n.t('SiliconCloud') },
  { value: 42, color: 'blue', label: i18n.t('Mistral AI') },
  { value: 8, color: 'pink', label: i18n.t('Custom channel') },
  {
    value: 22,
    color: 'blue',
    label: i18n.t('Knowledge base: FastGPT'),
  },
  {
    value: 21,
    color: 'purple',
    label: i18n.t('Knowledge base: AI Proxy'),
  },
  {
    value: 44,
    color: 'purple',
    label: i18n.t('Embedding: MokaAI M3E'),
  },
  {
    value: 45,
    color: 'blue',
    label: i18n.t('ByteDance Volcano Ark / Doubao / DeepSeek'),
  },
  {
    value: 48,
    color: 'blue',
    label: 'xAI'
  }
];
