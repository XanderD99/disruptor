# Contributing to This Project

Thank you for considering contributing to this Go project! Your help is highly appreciated.

Please take a moment to review this guide before submitting issues or pull requests.

---

## 🛠 Prerequisites

Before contributing, make sure you have the following installed:

- [Go](https://golang.org/dl/) ≥ 1.21
- [golangci-lint](https://golangci-lint.run/) (for linting)
- `make` (optional, if Makefile is provided)
- Git

---

## 🚀 Getting Started

1. Fork the repository  
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/your-repo.git
   cd your-repo
   ```
3. Create a new branch:
   ```bash
   git checkout -b feature/your-branch-name
   ```

---

## 📐 Code Style and Quality

- Format your code using `gofmt` or run:
  ```bash
  go fmt ./...
  ```

- Lint your code using `golangci-lint`:
  ```bash
  golangci-lint run
  ```

- Test your changes:
  ```bash
  go test ./...
  ```

- Run benchmarks if performance is relevant:
  ```bash
  go test -bench=. ./...
  ```

> Tip: You can use `make lint`, `make test`, etc. if a Makefile is provided.

---

## ✅ Pull Request Checklist

Before submitting a PR, make sure:

- [ ] Your code builds and passes all tests.
- [ ] You’ve written or updated relevant tests.
- [ ] `golangci-lint` passes with no errors.
- [ ] The code is formatted (`go fmt`).
- [ ] The PR has a clear title and description.
- [ ] Linked the related issue (if applicable).

---

## 🧪 Tests

We use Go's built-in testing framework. You can run all tests with:

```bash
go test ./...
```

If your change involves new behavior or fixes a bug, **write unit tests** that prove it works.

---

## 📄 Licensing

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thanks again for your help!
