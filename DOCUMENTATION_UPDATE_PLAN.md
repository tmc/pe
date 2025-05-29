# Documentation Update Plan

## Overview

This plan outlines the necessary updates to make PE's documentation accurate and reflect only implemented features.

## 1. Replace README.md

Replace the current README.md with README-accurate.md which:
- ✅ Only lists implemented commands
- ✅ Accurately describes the inference API with cgpt provider
- ✅ Correctly documents the plugin system
- ✅ Shows real working examples
- ✅ Removes all unimplemented features

## 2. Replace CLAUDE.md

Replace with CLAUDE-accurate.md which:
- ✅ Focuses on actual project architecture
- ✅ Lists only implemented commands
- ✅ Accurately describes the metaprompting implementation
- ✅ Documents the real plugin system
- ✅ Removes fictional features

## 3. Create ROADMAP.md

Move all unimplemented but planned features to a roadmap:

```markdown
# PE Roadmap

## Planned Features

### Go Toolchain Parity
- [ ] `pe build` - Optimize prompts for production
- [ ] `pe install` - Install prompt libraries
- [ ] `pe mod` - Dependency management
- [ ] `pe tool` - Generate prompt-based tools

### Advanced Features
- [ ] Gist and txtar format support
- [ ] macOS sandboxing and trust management
- [ ] History and session management
- [ ] Git-like version control
- [ ] Style guides and behavioral mixins
- [ ] Distributed caching with verification

### Provider Expansion
- [ ] Ollama provider
- [ ] OpenAI native provider
- [ ] Anthropic native provider
- [ ] Custom endpoint support
```

## 4. Update Documentation Files

### Files to Update:
1. **docs/QUICK_START.md** - Remove references to unimplemented commands
2. **docs/GETTING_STARTED.md** - Focus on actual capabilities
3. **docs/CLI_REFERENCE.md** - Already mostly accurate, minor updates needed
4. **docs/WORLD_CLASS_*.md** - Tone down claims, focus on implemented features

### Files that are Accurate:
- docs/ADVANCED_OPTIMIZATION_GUIDE.md
- docs/API_REFERENCE.md
- docs/ARCHITECTURE.md
- docs/METAPROMPTING_DESIGN_2025.md

## 5. Update Examples

### Keep These (Accurate):
- examples/run-command.sh
- examples/inference/basic.go
- examples/simple/
- examples/advanced-optimization/

### Review/Update:
- Any examples showing unimplemented commands

## 6. Key Messages to Preserve

PE still has significant value with:
1. **Cutting-edge optimization**: Semantic backpropagation, GASO, TextGrad
2. **Advanced metrics**: BLEU, ROUGE, BERTScore, G-Eval, UniEval
3. **Component composition**: Style-aware prompt engineering
4. **Plugin architecture**: Extensible design
5. **Unix philosophy**: Pipeline processing
6. **Go performance**: Fast, efficient, single binary

## 7. Implementation Steps

1. **Backup current docs**: Create a `docs-legacy/` folder
2. **Replace main files**: Use the accurate versions
3. **Create ROADMAP.md**: Document future plans
4. **Update docs/**: Fix individual documentation files
5. **Test all examples**: Ensure they work
6. **Update CI/CD**: Ensure tests reflect reality

## 8. Communication Strategy

When updating:
- Be honest about current capabilities
- Highlight the unique features that ARE implemented
- Position unimplemented features as "roadmap items"
- Focus on PE's real strengths:
  - Research implementation (semantic backprop, GASO)
  - Advanced metrics
  - Optimization methods
  - Plugin architecture
  - Go performance