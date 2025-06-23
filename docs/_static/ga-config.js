// Google Analytics configuration for Cano-Collector documentation
// Replace GA_MEASUREMENT_ID with your actual Google Analytics Measurement ID

const GA_CONFIG = {
    measurementId: 'GA_MEASUREMENT_ID', // Replace with your actual GA4 Measurement ID
    debugMode: false, // Set to true for development
    anonymizeIp: true,
    pageLoadTime: true,
    customDimensions: {
        pageType: 'documentation',
        project: 'cano-collector'
    }
};

// Initialize Google Analytics
function initGA() {
    if (typeof gtag !== 'undefined' && GA_CONFIG.measurementId !== 'GA_MEASUREMENT_ID') {
        gtag('config', GA_CONFIG.measurementId, {
            anonymize_ip: GA_CONFIG.anonymizeIp,
            custom_map: {
                'custom_map_1': 'page_type',
                'custom_map_2': 'project'
            }
        });
        
        // Set custom dimensions
        gtag('set', 'page_type', GA_CONFIG.customDimensions.pageType);
        gtag('set', 'project', GA_CONFIG.customDimensions.project);
    }
}

// Track custom events
function trackCustomEvent(eventName, parameters = {}) {
    if (typeof gtag !== 'undefined' && GA_CONFIG.measurementId !== 'GA_MEASUREMENT_ID') {
        gtag('event', eventName, {
            ...parameters,
            project: GA_CONFIG.customDimensions.project
        });
    }
}

// Track page load time
function trackPageLoadTime() {
    if (GA_CONFIG.pageLoadTime && typeof gtag !== 'undefined') {
        window.addEventListener('load', function() {
            const loadTime = performance.timing.loadEventEnd - performance.timing.navigationStart;
            trackCustomEvent('page_load_time', {
                value: loadTime,
                custom_parameter: 'load_time_ms'
            });
        });
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    initGA();
    trackPageLoadTime();
}); 