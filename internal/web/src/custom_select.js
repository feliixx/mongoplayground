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
 * @param {string} config.width - width of the custom select in px
 * @param {function} config.onChange - custom code to execute when an option is selected
 */
var CustomSelect = function (config) {

    const selectedClass = "is-selected"
    const openClass = "is-open"

    const select = document.getElementById(config.selectId)
    let selectedIndex = select.selectedIndex

    let options = []
    for (var i = 0; i < select.options.length; i++) {
        options.push(select.options[i].textContent)
    }

    const selectContainer = document.createElement("div")
    selectContainer.className = "js-Dropdown"
    selectContainer.id = "custom-" + select.id

    const button = document.createElement("button")
    button.className = "js-Dropdown-title"
    button.style.width = config.width

    const ul = document.createElement("ul")
    ul.className = "js-Dropdown-list"
    ul.style.width = config.width

    generateOptions()

    selectContainer.appendChild(button)
    selectContainer.appendChild(ul)
    selectContainer.addEventListener("click", onClick)

    // pseudo-select is ready - append it and remove the original
    select.parentNode.insertBefore(selectContainer, select)
    select.parentNode.removeChild(select)

    document.addEventListener("click", event => {
        if (!selectContainer.contains(event.target)) {
            close()
        }
    })

    function generateOptions() {

        button.textContent = ""

        for (var i = 0; i < options.length; i++) {

            var li = document.createElement("li")
            li.innerText = options[i]
            li.setAttribute("data-index", i)

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
            setSelectedIndex(event.target.getAttribute("data-index"))

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

    /**
     * 
     * @param {int} index - the index of the value to select 
     */
    function setSelectedIndex(index) {

        let liElem = ul.querySelectorAll("li")
        for (let i = 0; i < liElem.length; i++) {
            liElem[i].classList.remove(selectedClass)
            if (i == index) {
                selectedIndex = i
                liElem[i].classList.add(selectedClass)
                button.textContent = liElem[i].innerText
            }
        }
    }

    function getValue() {
        return button.textContent
    }

    /**
     * Set the select value from a string. Caution, do not use this 
     * method if the select contains the same option several times 
     * 
     * @param {string} value - the option to select 
     */
    function setValue(value) {

        button.textContent = value

        let liElem = ul.querySelectorAll("li")
        for (let i = 0; i < liElem.length; i++) {
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