#!/usr/bin/env python3
# -*- coding: utf-8 -*-
#
# Cano-Collector documentation build configuration file
#
# This file is execfile()d with the current directory set to its
# containing dir.
#
# Note that not all possible configuration values are present in this
# autogenerated file.
#
# All configuration values have a default; values that are commented out
# serve to show the default.

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
#
import sys
from pathlib import Path

from docutils import nodes

# Add the root cano-collector directory to the path
sys.path.insert(0, str(Path(__file__).parent))
sys.path.insert(0, str(Path(__file__).parent.parent))

# -- General configuration ------------------------------------------------

extensions = [
    "sphinx.ext.autodoc",
    "sphinx.ext.graphviz",
    "sphinx.ext.inheritance_diagram",
    "sphinx.ext.autosummary",
    "sphinxcontrib.mermaid",
    "sphinx.ext.napoleon",
    "sphinx_autodoc_typehints",
    "sphinx.ext.autosectionlabel",
    "sphinx.ext.viewcode",
    "sphinx_design",
    "sphinxcontrib.images",
    "sphinx_immaterial",
    "sphinx_jinja",
    "sphinx_reredirects",
]

suppress_warnings = ["autosectionlabel.*"]

# for sphinx_jinja
jinja_contexts = {}

jinja_filters = {
    "to_snake_case": lambda s: "".join(
        ["_" + c.lower() if c.isupper() else c for c in s]
    ).lstrip("_")
}

images_config = {
    "override_image_directive": True,
}

smartquotes = False

autodoc_mock_imports = []
autodoc_default_options = {
    "members": True,
    "member-order": "bysource",
    "undoc-members": True,
}
autoclass_content = "both"
add_module_names = False

# Add any paths that contain templates here, relative to this directory.
templates_path = ["_templates"]

# The suffix(es) of source filenames.
source_suffix = [".rst", ".md"]

# The master toctree document.
master_doc = "index"

# General information about the project.
project = "Cano-Collector"
copyright = "2025, Kubecano"
author = "Kubecano"

# The short X.Y version.
# version = "DOCS_VERSION_PLACEHOLDER"
# The full version, including alpha/beta/rc tags.
# release = "DOCS_RELEASE_PLACEHOLDER"

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
exclude_patterns = ["_build", "Thumbs.db", ".DS_Store", "**/*.inc.rst", "**/*.jinja", "venv", "README.md"]

# The name of the Pygments (syntax highlighting) style to use.
pygments_style = "witchhazel.WitchHazelStyle"
pygments_dark_style = "witchhazel.WitchHazelStyle"

# If true, `todo` and `todoList` produce output, else they produce nothing.
todo_include_todos = False

html_theme = "sphinx_immaterial"

html_logo = "_static/logo_default.svg"

html_theme_options = {
    "icon": {
        "repo": "fontawesome/brands/github",
    },
    "repo_url": "https://github.com/kubecano/cano-collector/",
    "repo_name": "Cano-Collector",
    "edit_uri": "tree/master/docs",
    "palette": [
        {
            "media": "(prefers-color-scheme: light)",
            "scheme": "default",
            "primary": "cano",
            "accent": "cano",
            "toggle": {
                "icon": "material/weather-night",
                "name": "Switch to dark mode",
            },
        },
        {
            "media": "(prefers-color-scheme: dark)",
            "scheme": "slate",
            "primary": "cano-dark",
            "accent": "cano-dark",
            "toggle": {
                "icon": "material/weather-sunny",
                "name": "Switch to light mode",
            },
        },
    ],
    "features": [
        "navigation.instant",
        "navigation.top",
        "navigation.tabs",
        "navigation.tabs.sticky",
        "search.share",
        "toc.follow",
        "toc.sticky",
    ],
    "globaltoc_collapse": False,
    "social": [
        {
            "icon": "fontawesome/brands/github",
            "link": "https://github.com/kubecano/cano-collector/",
        },
    ],
}

html_sidebars = {
    "**": ["logo-text.html", "globaltoc.html", "localtoc.html", "searchbox.html"]
}

copybutton_prompt_text = r"$ "

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
html_static_path = ["_static"]

html_css_files = [
    "custom.css",
    "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.1.1/css/all.min.css",
]

html_js_files = ["analytics.js", "ga-config.js"]

html_favicon = "_static/favicon.png"

html_extra_path = ["robots.txt"]


def setup(app):
    app.add_css_file("custom.css")
    app.add_role('checkmark', checkmark_role)

def checkmark_role(name, rawtext, text, lineno, inliner, options=None, content=None):
    if options is None:
        options = {}
    if content is None:
        content = []
    node = nodes.inline(text='✓', classes=['success-icon'])
    return [node], []
