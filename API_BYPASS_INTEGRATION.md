# API Bypass Integration Guide

## Overview
This guide explains how to enable API bypass functionality for AI models in the nofx system. The bypass allows using browser automation instead of direct API calls, helping to avoid API costs and rate limits.

## How It Works
- The system uses a hook-based approach to intercept API calls
- When bypass is enabled, API calls are redirected to browser automation
- All other parts of the nofx flow remain unchanged

## Enabling API Bypass

To enable API bypass for any AI model, set either of these configuration options:

1. Set `CustomModelName` to `"bypass"`
2. Include `"bypass"` in the `CustomAPIURL` field

### Examples:

**In trader configuration:**
```go
config := AutoTraderConfig{
    AIModel: "deepseek",           // Or any other model
    CustomModelName: "bypass",    // Enable bypass
    // Other settings...
}
```

**Or using CustomAPIURL:**
```go
config := AutoTraderConfig{
    AIModel: "openai",                         // Or any other model
    CustomAPIURL: "https://api.example.com?bypass=true", // Include "bypass" in the URL
    // Other settings...
}
```

## Supported Models
All nofx AI models support the bypass functionality:
- DeepSeek (`deepseek`)
- OpenAI (`openai`) 
- Claude (`claude`)
- Qwen (`qwen`)
- Google Gemini (`gemini`)
- Grok (`grok`)
- Kimi (`kimi`)
- Ollama (`ollama`)
- Guardian (`guardian`)
- Custom (`custom`)

## Browser Automation Providers
When bypass is enabled, the system uses appropriate browser automation providers:
- `deepseek-browser` for DeepSeek
- `chatgpt-browser` for OpenAI/ChatGPT
- `claude-browser` for Claude
- `guardian-ai` for other models (as fallback)

## Implementation Details
- The `BypassClient` wraps the original AI client
- Uses the existing nofx hook system for method interception
- Zero changes to original nofx codebase
- Maintains full compatibility with existing functionality