var YlaTooltip = function () {
    // Private method
    function _createTooltip() {
        var tooltip = document.createElement('span');
        tooltip.className = 'yla-tooltip';
        return tooltip;
    }

    function _showTooltip(target, tooltip, placement) {
        // Put tooltip in document
        document.body.appendChild(tooltip);

        // Calculate position for tooltip
        var placement = placement ? placement + '' : '',
            targetRect = target.getBoundingClientRect(),
            tooltipRect = tooltip.getBoundingClientRect(),
            targetCenterX = targetRect.left + (targetRect.width / 2),
            targetCenterY = targetRect.top + (targetRect.height / 2),
            className, tooltipX, tooltipY;

        switch (placement.toLowerCase()) {
            case 'top':
                className = 'top';
                tooltipX = targetCenterX - (tooltipRect.width / 2);
                tooltipY = targetRect.top - tooltipRect.height;
                break;
            case 'bottom':
                className = 'bottom';
                tooltipX = targetCenterX - (tooltipRect.width / 2);
                tooltipY = targetRect.bottom;
                break;
            case 'left':
                className = 'left';
                tooltipX = targetRect.left - tooltipRect.width;
                tooltipY = targetCenterY - (tooltipRect.height / 2);
                break;
            case 'right':
            default:
                className = 'right';
                tooltipX = targetRect.right;
                tooltipY = targetCenterY - (tooltipRect.height / 2);
                break;
        }

        // Position tooltip
        tooltip.style.position = 'fixed';
        tooltip.style.top = tooltipY + 'px';
        tooltip.style.left = tooltipX + 'px';
        tooltip.className = 'yla-tooltip ' + className;
    }

    function _removeTooltip(tooltip) {
        document.body.removeChild(tooltip);
    }

    return {
        props: {
            placement: {
                type: String,
                default: ''
            },
            content: {
                type: String,
                default: ''
            }
        },
        data: function () {
            return {
                tooltip: _createTooltip()
            };
        },
        watch: {
            content: {
                immediate: true,
                handler: function () {
                    this.tooltip.textContent = this.content;
                }
            }
        },
        render: function (createElement) {
            // Make sure this component contain at least one element
            var nodes = this.$slots.default || [],
                mainElement = nodes.find(node => {
                    return node.tag && node.tag !== '';
                });

            if (!mainElement) return;

            // Set event handler for main element
            var newData = mainElement.data || {};

            newData.on = newData.on || {};
            newData.on.mouseenter = (evt) => {
                _showTooltip(evt.target, this.tooltip, this.placement);
            };

            newData.on.mouseleave = () => {
                _removeTooltip(this.tooltip);
            };

            // Return main element
            mainElement.data = newData;
            return mainElement;
        }
    }
};