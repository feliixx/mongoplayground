/**
 * adapatation from https://github.com/zoltantothcom/vanilla-js-dropdown by Zoltan Toth
 *
 * @class
 * @param {(string)} options.elem - HTML id of the select.
 */
var CustomSelect = function (options) {

    var elem = document.getElementById(options.elem)
    var mainClass = 'js-Dropdown'
    var titleClass = 'js-Dropdown-title'
    var listClass = 'js-Dropdown-list'
    var selectedClass = 'is-selected'
    var openClass = 'is-open'
    var selectOptions = elem.options
    var optionsLength = selectOptions.length
    var index = 0

    var selectContainer = document.createElement('div')
    selectContainer.className = mainClass
    selectContainer.id = 'custom-' + elem.id

    var button = document.createElement('button')
    button.className = titleClass
    button.textContent = selectOptions[0].textContent

    var ul = document.createElement('ul')
    ul.className = listClass

    generateOptions(selectOptions)

    selectContainer.appendChild(button)
    selectContainer.appendChild(ul)
    selectContainer.addEventListener('click', onClick)

    // pseudo-select is ready - append it and hide the original
    elem.parentNode.insertBefore(selectContainer, elem)
    elem.style.display = 'none'

    function generateOptions(options) {
        for (var i = 0; i < options.length; i++) {

            var li = document.createElement('li')
            li.innerText = options[i].textContent
            li.setAttribute('data-value', options[i].value)
            li.setAttribute('data-index', index++)

            if (selectOptions[elem.selectedIndex].textContent === options[i].textContent) {
                li.classList.add(selectedClass)
                button.textContent = options[i].textContent
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
            elem.options.selectedIndex = t.getAttribute('data-index')
            elem.dispatchEvent(new CustomEvent('change'))

            var liElem = ul.querySelectorAll('li')
            for (var i = 0; i < optionsLength; i++) {
                liElem[i].classList.remove(selectedClass)
            }
            t.classList.add(selectedClass)
            close()
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

    function getValue() {
        return selectContainer.querySelector('.' + titleClass).innerText
    }

    function setValue(value) {
        var liElem = ul.querySelectorAll('li')
        for (var i = 0; i < optionsLength; i++) {
            liElem[i].classList.remove(selectedClass)
            if (liElem[i].innerText == value) {
                liElem[i].classList.add(selectedClass)
            }
        }
        selectContainer.querySelector('.' + titleClass).innerText = value
    }

    return {
        toggle: toggle,
        close: close,
        open: open,
        getValue: getValue,
        setValue: setValue
    }
}