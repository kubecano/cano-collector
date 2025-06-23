.. cano-collector documentation master file, created by
   sphinx-quickstart on Thu Jun 20 15:47:49 2024.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

.. meta::
   :title: Cano-Collector Documentation
   :description: Cano-collector is a Kubernetes-native alert processing and notification system.
   :author: Kubecano
   :robots: index, follow
   :og:title: Cano-Collector Documentation
   :og:description: Cano-collector is a Kubernetes-native alert processing and notification system.
   :og:type: website
   :og:site_name: Cano-Collector Docs

Welcome to cano-collector's documentation!
============================================

Cano-collector is a Kubernetes-native alert processing and notification system that receives Prometheus alerts from Alertmanager and routes them to various notification destinations like Slack, MS Teams, Jira, and more.

.. toctree::
   :hidden:
   :maxdepth: 2
   :caption: Getting Started

   quick_start
   installation
   first_alert

.. toctree::
   :hidden:
   :maxdepth: 2
   :caption: User Guide

   architecture/index
   configuration/index

.. toctree::
   :hidden:
   :maxdepth: 2
   :caption: Developer Guide

   development_guide
   api_reference
   implementation_tasks

.. toctree::
   :hidden:
   :maxdepth: 2
   :caption: Operations

   troubleshooting
   monitoring
   maintenance

Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`