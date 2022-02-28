/**
 * adapatation from https://github.com/zoltantothcom/vanilla-js-dropdown by Zoltan Toth
 *
 * @class
 * @param {(string)} config.selectId - HTML id of the select.
 * @param {(function)} config.onChange - custom code to execute when an option is selected
 */
var CustomSelect = function (config) {

    var select = document.getElementById(config.selectId)
    var mainClass = 'js-Dropdown'
    var titleClass = 'js-Dropdown-title'
    var listClass = 'js-Dropdown-list'
    var selectedClass = 'is-selected'
    var openClass = 'is-open'

    var selectContainer = document.createElement('div')
    selectContainer.className = mainClass
    selectContainer.id = 'custom-' + select.id

    var button = document.createElement('button')
    button.className = titleClass

    var ul = document.createElement('ul')
    ul.className = listClass

    generateOptions(select.options)

    selectContainer.appendChild(button)
    selectContainer.appendChild(ul)
    selectContainer.addEventListener('click', onClick)

    // pseudo-select is ready - append it and hide the original
    select.parentNode.insertBefore(selectContainer, select)
    select.style.display = 'none'

    function generateOptions(options) {

        for (var i = 0; i < options.length; i++) {

            var li = document.createElement('li')
            li.innerText = options[i].textContent
            li.setAttribute('data-value', options[i].value)
            li.setAttribute('data-index', i)

            if (i === select.selectedIndex) {
                li.classList.add(selectedClass)
                button.textContent = li.innerText
            }

            ul.appendChild(li)
        }
    }

    document.addEventListener('click', function (e) {
        if (!selectContainer.contains(e.target)) {
            close()
        }
    })

    function onClick(e) {
        e.preventDefault()

        var t = e.target

        if (t.className === titleClass) {
            toggle()
        }

        if (t.tagName === 'LI') {

            selectContainer.querySelector('.' + titleClass).innerText = t.innerText
            select.options.selectedIndex = t.getAttribute('data-index')
            select.dispatchEvent(new CustomEvent('change'))

            var liElem = ul.querySelectorAll('li')
            for (var i = 0; i < liElem.length; i++) {
                liElem[i].classList.remove(selectedClass)
            }
            t.classList.add(selectedClass)
            close()

            if (typeof config.onChange === 'function') {
                config.onChange()
            }
        }
    }

    function toggle() {
        ul.classList.toggle(openClass)
    }

    function open() {
        ul.classList.add(openClass)
    }

    function close() {
        ul.classList.remove(openClass)
    }

    function getSelectedIndex() {
        if (select.options.length === 0) {
            return -1
        }
        return select.selectedIndex
    }

    function getValue() {
        return selectContainer.querySelector('.' + titleClass).innerText
    }

    function setValue(value) {
        var liElem = ul.querySelectorAll('li')
        for (var i = 0; i < liElem.length; i++) {
            liElem[i].classList.remove(selectedClass)
            if (liElem[i].innerText == value) {
                liElem[i].classList.add(selectedClass)
            }
        }
        selectContainer.querySelector('.' + titleClass).innerText = value
    }

    function setOptions(optionMap) {

        while (ul.firstChild) {
            ul.firstChild.remove()
            select.firstChild.remove()
        }
        var size = 0
        for (var key in optionMap) {
            var opt = document.createElement('option')
            opt.textContent = key
            opt.value = optionMap[key]

            select.appendChild(opt)
            size++
        }

        if (size > 0) {
            select.selectedIndex = size - 1
        }
        generateOptions(select.options)
    }

    return {
        toggle: toggle,
        close: close,
        open: open,
        getSelectedIndex: getSelectedIndex,
        getValue: getValue,
        setValue: setValue,
        setOptions: setOptions
    }
}