/* ZeroPass Landing Page — JavaScript */

(function () {
    'use strict';

    // ── Scroll-based reveal animations ──
    function initScrollReveal() {
        const targets = document.querySelectorAll(
            '.bento-card, .security-card, .step'
        );
        if (!targets.length) return;

        const prefersReduced = window.matchMedia(
            '(prefers-reduced-motion: reduce)'
        ).matches;
        if (prefersReduced) {
            targets.forEach(function (el) {
                el.classList.add('visible');
            });
            return;
        }

        var observer = new IntersectionObserver(
            function (entries) {
                entries.forEach(function (entry) {
                    if (entry.isIntersecting) {
                        entry.target.classList.add('visible');
                        observer.unobserve(entry.target);
                    }
                });
            },
            { threshold: 0.15, rootMargin: '0px 0px -40px 0px' }
        );

        targets.forEach(function (el, i) {
            el.style.transitionDelay = (i % 3) * 100 + 'ms';
            observer.observe(el);
        });
    }

    // ── Navbar scroll effect ──
    function initNavbar() {
        var navbar = document.getElementById('navbar');
        if (!navbar) return;

        var scrolled = false;
        window.addEventListener(
            'scroll',
            function () {
                var isScrolled = window.scrollY > 10;
                if (isScrolled !== scrolled) {
                    scrolled = isScrolled;
                    navbar.style.background = scrolled
                        ? 'rgba(10, 15, 26, 0.95)'
                        : 'rgba(10, 15, 26, 0.8)';
                }
            },
            { passive: true }
        );
    }

    // ── Mobile menu toggle ──
    function initMobileMenu() {
        var toggle = document.getElementById('mobileToggle');
        var links = document.querySelector('.nav-links');
        if (!toggle || !links) return;

        toggle.addEventListener('click', function () {
            var isOpen = links.style.display === 'flex';
            links.style.display = isOpen ? 'none' : 'flex';
            links.style.flexDirection = 'column';
            links.style.position = 'absolute';
            links.style.top = '64px';
            links.style.left = '0';
            links.style.right = '0';
            links.style.background = 'rgba(10, 15, 26, 0.98)';
            links.style.padding = isOpen ? '0' : '16px 24px';
            links.style.gap = '16px';
            links.style.borderBottom = isOpen
                ? 'none'
                : '1px solid rgba(255,255,255,0.08)';
        });
    }

    // ── Copy to clipboard buttons ──
    function initCopyButtons() {
        document.querySelectorAll('.copy-btn').forEach(function (btn) {
            btn.addEventListener('click', function () {
                var code = btn.getAttribute('data-code');
                if (!code) return;

                navigator.clipboard
                    .writeText(code)
                    .then(function () {
                        var original = btn.innerHTML;
                        btn.innerHTML =
                            '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#00E5A0" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>';
                        setTimeout(function () {
                            btn.innerHTML = original;
                        }, 2000);
                    })
                    .catch(function () {
                        /* clipboard not available */
                    });
            });
        });
    }

    // ── Terminal typing effect ──
    function initTerminalAnimation() {
        var body = document.querySelector('.terminal-body');
        if (!body) return;

        var lines = body.querySelectorAll('.terminal-line');
        lines.forEach(function (line, i) {
            line.style.opacity = '0';
            line.style.transform = 'translateY(4px)';
        });

        var observer = new IntersectionObserver(
            function (entries) {
                entries.forEach(function (entry) {
                    if (entry.isIntersecting) {
                        lines.forEach(function (line, i) {
                            setTimeout(function () {
                                line.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
                                line.style.opacity = '1';
                                line.style.transform = 'translateY(0)';
                            }, i * 80);
                        });
                        observer.unobserve(entry.target);
                    }
                });
            },
            { threshold: 0.3 }
        );

        observer.observe(body);
    }

    // ── Smooth scroll for anchor links ──
    function initSmoothScroll() {
        document.querySelectorAll('a[href^="#"]').forEach(function (link) {
            link.addEventListener('click', function (e) {
                var target = document.querySelector(link.getAttribute('href'));
                if (target) {
                    e.preventDefault();
                    target.scrollIntoView({ behavior: 'smooth', block: 'start' });

                    // Close mobile menu if open
                    var navLinks = document.querySelector('.nav-links');
                    if (window.innerWidth <= 768 && navLinks) {
                        navLinks.style.display = 'none';
                    }
                }
            });
        });
    }

    // ── Init all ──
    document.addEventListener('DOMContentLoaded', function () {
        initNavbar();
        initMobileMenu();
        initScrollReveal();
        initCopyButtons();
        initTerminalAnimation();
        initSmoothScroll();
    });
})();
