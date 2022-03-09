/* ***** BEGIN LICENSE BLOCK *****
* A simple dropdown in vanilla js
*
* Copyright (C) 2015 Zoltan Toth
* Copyright (C) 2017 Adrien Petel
*
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU Affero General Public License as published
* by the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
* GNU Affero General Public License for more details.
*
* You should have received a copy of the GNU Affero General Public License
* along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * ***** END LICENSE BLOCK ***** */

/**
 * A simple dropdown in vanilla js. Adaptation from https://github.com/zoltantothcom/vanilla-js-dropdown
 * by Zoltan Toth
 *
 * @class
 * @param {string} config.selectId - HTML id of the select.
 * @param {function} config.onChange - custom code to execute when an option is selected
 * @param {string} config.width - the desired width of the elem, in px, like "120px"
 */
var CustomSelect = function (config) {

    var selectedClass = "is-selected"
    var openClass = "is-open"

    var select = document.getElementById(config.selectId)
    var selectedIndex = select.selectedIndex

    var options = []
    for (var i = 0; i < select.options.length; i++) {
        options.push(select.options[i].textContent)
    }

    var selectContainer = document.createElement("div")
    selectContainer.className = "js-Dropdown"
    selectContainer.id = "custom-" + select.id

    var button = document.createElement("button")
    button.className = "js-Dropdown-title"
    button.style.width = config.width

    var ul = document.createElement("ul")
    ul.className = "js-Dropdown-list"
    ul.style.width = config.width

    generateOptions()

    selectContainer.appendChild(button)
    selectContainer.appendChild(ul)
    selectContainer.addEventListener("click", onClick)

    // pseudo-select is ready - append it and remove the original
    select.parentNode.insertBefore(selectContainer, select)
    select.parentNode.removeChild(select)

    document.addEventListener("click", function (e) {
        if (!selectContainer.contains(e.target)) {
            close()
        }
    })

    function generateOptions() {

        button.textContent = ""

        for (var i = 0; i < options.length; i++) {

            var li = document.createElement("li")
            li.innerText = options[i]

            if (i === selectedIndex) {
                li.classList.add(selectedClass)
                button.textContent = li.innerText
            }
            ul.appendChild(li)
        }
    }

    function onClick(event) {

        event.preventDefault()

        if (event.target.tagName === "LI") {
            setValue(event.target.innerText)

            if (typeof config.onChange === "function") {
                config.onChange()
            }
        }
        toggle()
    }

    function toggle() {
        ul.classList.toggle(openClass)
    }

    function close() {
        ul.classList.remove(openClass)
    }

    function getSelectedIndex() {
        return selectedIndex
    }

    function getValue() {
        return button.textContent
    }

    /**
     * 
     * @param {string} value - the option to select 
     */
    function setValue(value) {

        button.textContent = value

        var liElem = ul.querySelectorAll("li")
        for (var i = 0; i < liElem.length; i++) {
            liElem[i].classList.remove(selectedClass)
            if (liElem[i].innerText == value) {
                selectedIndex = i
                liElem[i].classList.add(selectedClass)
            }
        }
    }

    /**
     * Set the options for the select. The last option is selected
     * 
     * @param {string[]} optionList - a list of string
     */
    function setOptions(optionList) {

        while (ul.firstChild) {
            ul.firstChild.remove()
        }
        options = optionList

        if (options.length === 0) {
            selectedIndex = -1
        } else {
            selectedIndex = options.length - 1
        }
        generateOptions()
    }

    return {
        toggle: toggle,
        getSelectedIndex: getSelectedIndex,
        getValue: getValue,
        setValue: setValue,
        setOptions: setOptions
    }
}