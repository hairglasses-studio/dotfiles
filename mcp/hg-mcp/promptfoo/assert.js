module.exports = {
  validateOllamaOutput: (output, context) => {
    const vars = context?.vars || {};
    const checks = [];
    let parsed;

    try {
      parsed = JSON.parse(output);
      checks.push({
        pass: true,
        score: 1,
        reason: 'Output is valid JSON',
      });
    } catch (e) {
      return {
        pass: false,
        score: 0,
        reason: 'Output is not valid JSON',
      };
    }

    if (vars.tool_name === 'aftrs_ollama_structured') {
      const status = parsed.parsed_json?.status;
      const pass = status === 'OK';
      checks.push({
        pass,
        score: pass ? 1 : 0,
        reason: `Expected status=OK, got ${status}`,
      });
    }

    if (vars.tool_name === 'aftrs_ollama_tool_chat') {
      const toolCalls = parsed.tool_calls;
      const pass = Array.isArray(toolCalls) && toolCalls.length > 0 && toolCalls[0].function?.name === 'test_tool';
      checks.push({
        pass,
        score: pass ? 1 : 0,
        reason: pass ? 'Tool call returned correctly' : 'Expected tool call not returned',
      });
    }

    const pass = checks.every((check) => check.pass);
    const score = checks.reduce((sum, check) => sum + check.score, 0) / checks.length;
    
    return {
      pass,
      score,
      reason: pass ? 'Ollama tool response valid' : 'Ollama tool response regression',
      componentResults: checks,
    };
  }
};