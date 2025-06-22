# Cano-Collector Documentation

This directory contains documentation for the Cano-Collector project.

## Building Documentation

### Requirements

- Python 3.8+
- pip

### Installation with virtual environment (recommended for local development)

```bash
# Create virtual environment
python -m venv venv

# Activate virtual environment
# On macOS/Linux:
source venv/bin/activate
# On Windows:
venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt
```

### Building HTML

```bash
make html
```

### Serving locally

```bash
make serve
```

Documentation will be available at: http://localhost:8000

### Building with Docker

```bash
make docker-build
```

### Other useful commands

- `make clean` - clean build directory
- `make linkcheck` - check links
- `make html-strict` - build with warnings as errors
- `make deploy` - build for deployment (clean and build)
- `make local-setup` - setup local environment with venv

## File Structure

- `conf.py` - Sphinx configuration
- `index.rst` - documentation homepage
- `_static/` - static files (CSS, JS, images)
- `_templates/` - HTML templates
- `requirements.txt` - Python dependencies

## Google Analytics Configuration

To enable Google Analytics:

1. Replace `GA_MEASUREMENT_ID` in `_static/ga-config.js` with your actual ID
2. Update `_templates/base.html` with the actual ID

## Deployment

The documentation will be available at: https://kubecano.com/docs/cano-collector

After building documentation (`make deploy`), HTML files will be in the `_build/html/` directory.

You can deploy to S3 using AWS CLI:

```bash
aws s3 sync _build/html/ s3://your-bucket-name/ --delete
```

## Theme

Documentation uses the `sphinx_immaterial` theme with custom CSS styles. 