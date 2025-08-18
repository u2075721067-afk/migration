# Contributing Guidelines

Thank you for your interest in contributing to this project! This document outlines the standards and practices we follow to maintain code quality and consistency.

## Code Style

### Python
- Follow **PEP 8** style guidelines
- Use meaningful variable and function names
- Keep functions focused and concise
- Add type hints where appropriate
- **All comments and documentation must be in English**

### General
- Use consistent indentation (4 spaces for Python)
- Follow the existing code structure and patterns
- Write self-documenting code when possible

## Git Workflow

### Commit Messages
Use the following format for all commit messages:
```
<type>(scope): <message>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(context): add new ctx-save target to Makefile
fix(db): resolve database connection timeout issue
docs(api): update API endpoint documentation
```

### Branching Strategy
- **`main`**: Production-ready code
- **`developer`**: Development branch for active work
- Create feature branches from `developer` when needed
- Submit Pull Requests to merge into `developer`
- `developer` gets merged to `main` after testing

### Commit Guidelines
- **One microtask per commit** - keep commits focused and atomic
- Commit frequently with clear, descriptive messages
- Test your changes before committing
- Use `git add -p` to review changes before committing

## Context Management

### After Each Completed Step
Always use the context save command after completing a significant step:
```bash
make ctx-save TITLE="Step Description" TEXT="What was accomplished" TAGS="relevant,tags"
```

This helps maintain project history and context for future development.

## Language Rules

### Communication
- **Chat and discussions**: Ukrainian
- **All code, documentation, and commit messages**: English only

### Why This Rule?
- Code and documentation in English ensures international accessibility
- Ukrainian communication maintains local team collaboration
- Consistent language separation prevents confusion

## Development Process

1. **Plan**: Understand the task and break it into microtasks
2. **Implement**: Write code following style guidelines
3. **Test**: Verify your changes work as expected
4. **Commit**: Use proper commit message format
5. **Context**: Save context with `make ctx-save`
6. **Review**: Self-review before submitting PR

## Getting Help

- Check existing documentation first
- Use `make help` to see available context utilities
- Ask questions in Ukrainian for clarity
- Provide context when reporting issues

## Code Review

- All code must be reviewed before merging
- Reviewers should check:
  - Code style compliance
  - Functionality correctness
  - Test coverage
  - Documentation updates
  - Context preservation

## Test Coverage Requirements

### Minimum Coverage Threshold
- **Tests must maintain at least 70% coverage**
- CI will fail if coverage drops below this threshold
- Coverage reports are generated for every test run
- HTML coverage reports are available as CI artifacts

### Coverage Best Practices
- Write tests for new functionality
- Ensure edge cases are covered
- Test error handling paths
- Maintain or improve coverage with each change
- Use `pytest --cov=ctx --cov-report=term-missing` locally to check coverage

Thank you for contributing to our project!
