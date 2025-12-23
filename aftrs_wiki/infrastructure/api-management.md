# API Key Management

## Overview
AFTRS ecosystem maintains default API keys for various LLM providers to enable AI-powered functionality across all tools and services.

## Default API Keys

### Anthropic (Claude)
- **Key**: Stored in `default_api_keys.json`
- **Models Available**:
  - claude-3-5-sonnet-latest
  - claude-3-5-haiku-latest
  - claude-3-opus-20240229
- **Status**: ✅ Working (as of Sept 2025)

### OpenAI (GPT)
- **Key**: Stored in `default_api_keys.json`
- **Models Available**:
  - gpt-4-turbo
  - gpt-4o
  - gpt-4o-mini
  - gpt-3.5-turbo
- **Status**: ✅ Working (as of Sept 2025)

## Testing Tools

### Location
- Primary: `/Users/mitch/Docs/aftrs-void/test_multi_api_keys.py`
- Anthropic-specific: `/Users/mitch/Docs/aftrs-void/test_anthropic_keys.py`
- AFTRS CLI integration: `/Users/mitch/Docs/aftrs-void/aftrs_cli/scripts/test_multi_api_keys.py`

### Usage
```bash
# Test both providers
python3 test_multi_api_keys.py

# Test via AFTRS CLI
./aftrs.sh api test multi
```

### Output Files
- `api_keys_test_results.json` - Detailed test results
- `default_api_keys.json` - Stored configuration
- `anthropic_key_test_results.json` - Anthropic-specific results

## Integration Points

### Cline (VS Code Extension)
- Anthropic key can be added directly in Cline settings
- OpenAI requires proxy through LiteLLM or OpenRouter

### LiteLLM Configuration
Located at: `aftrs-code-llm-plan/configs/litellm.config.yaml`
- Supports both Anthropic and OpenAI models
- Provides unified interface for multiple providers

### Docker Services
- `aftrs-code-llm-plan/docker-compose.yml` - LiteLLM proxy service
- `aftrs-code-llm-plan/docker-compose.claude-direct.yml` - Direct Claude access

## Security Considerations

### Storage
- API keys are stored locally in JSON configuration files
- Never commit API keys to public repositories
- Use environment variables for production deployments

### Access Control
- Keys are scoped to specific models and capabilities
- Monitor usage through provider dashboards
- Rotate keys regularly for security

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify key is correctly formatted
   - Check for extra spaces or characters
   - Ensure key hasn't been revoked

2. **Insufficient Credits**
   - Add credits via provider dashboard
   - Check usage limits and quotas
   - Consider upgrading plan if needed

3. **Model Access Issues**
   - Some models require specific access levels
   - Check deprecation notices for older models
   - Verify model names are correct

### Testing Workflow
1. Run `test_multi_api_keys.py` to verify keys
2. Check output for accessible models
3. Review `api_keys_test_results.json` for details
4. Update `default_api_keys.json` if needed

## Related Documentation
- [Enhanced AFTRS CLI Capabilities](../projects/aftrs_cli.md)
- [LLM Workflows Documentation](../projects/aftrs-code-llm-plan.md)
- [Git Identity Management](../infrastructure/git-identity.md)
