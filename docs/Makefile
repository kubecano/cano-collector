# Makefile for Sphinx documentation
#

# You can set these variables from the command line.
SPHINXOPTS    =
SPHINXBUILD   = sphinx-build
SPHINXPROJ    = cano-collector
SOURCEDIR     = .
BUILDDIR      = _build

# Put it first so that "make" without argument is like "make help".
help:
	@$(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

# Setup local environment with virtual environment
local-setup:
	@echo "Setting up local environment with virtual environment..."
	python -m venv venv
	@echo "Virtual environment created. Please activate it:"
	@echo "  source venv/bin/activate  # On macOS/Linux"
	@echo "  venv\\Scripts\\activate     # On Windows"
	@echo "Then run: pip install -r requirements.txt"

# Install dependencies in virtual environment
install-venv:
	@echo "Installing dependencies in virtual environment..."
	venv/bin/pip install -r requirements.txt

# Build HTML documentation
html:
	@$(SPHINXBUILD) -b html "$(SOURCEDIR)" "$(BUILDDIR)/html" $(SPHINXOPTS) $(O)

# Build HTML documentation with warnings as errors
html-strict:
	@$(SPHINXBUILD) -b html -W "$(SOURCEDIR)" "$(BUILDDIR)/html" $(SPHINXOPTS) $(O)

# Build HTML documentation using virtual environment
html-venv:
	@echo "Building documentation using virtual environment..."
	venv/bin/$(SPHINXBUILD) -b html "$(SOURCEDIR)" "$(BUILDDIR)/html" $(SPHINXOPTS) $(O)

# Build HTML documentation using Docker
docker-build:
	@echo "Building documentation using Docker..."
	docker run --rm -v "$(PWD):/docs" sphinxdoc/sphinx make html

# Clean build directory
clean:
	rm -rf "$(BUILDDIR)"

# Clean virtual environment
clean-venv:
	rm -rf venv

# Install dependencies
install:
	pip install -r requirements.txt

# Build and serve documentation locally
serve: html
	cd "$(BUILDDIR)/html" && python -m http.server 8000

# Build and serve documentation locally using virtual environment
serve-venv: html-venv
	cd "$(BUILDDIR)/html" && python -m http.server 8000

# Build documentation for deployment
deploy: clean html

# Build documentation for deployment using virtual environment
deploy-venv: clean html-venv

# Check for broken links
linkcheck:
	@$(SPHINXBUILD) -b linkcheck "$(SOURCEDIR)" "$(BUILDDIR)/linkcheck" $(SPHINXOPTS) $(O)

# Check for broken links using virtual environment
linkcheck-venv:
	@venv/bin/$(SPHINXBUILD) -b linkcheck "$(SOURCEDIR)" "$(BUILDDIR)/linkcheck" $(SPHINXOPTS) $(O)

# Build PDF documentation
pdf:
	@$(SPHINXBUILD) -b latex "$(SOURCEDIR)" "$(BUILDDIR)/latex" $(SPHINXOPTS) $(O)
	@echo "Running LaTeX files through pdflatex..."
	$(MAKE) -C "$(BUILDDIR)/latex" all-pdf
	@echo "pdflatex finished; the PDF files are in $(BUILDDIR)/latex."

.PHONY: help Makefile html html-strict html-venv docker-build clean clean-venv install install-venv serve serve-venv deploy deploy-venv linkcheck linkcheck-venv pdf

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: Makefile
	@$(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O) 