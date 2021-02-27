/* This JavaScript code uses ECMAScript 2015 (ES6) */
/* http://es6-features.org/ */

/* https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Strict_mode */
'use strict';

function _query(query) {
    return document.querySelector(query);
}

function _queryAll(query) {
    return document.querySelectorAll(query);
}


function tooltips() {
    // all elements with a "data-tip" attribute
    const selector = "a[data-tip]";

    // function that handles onclick events
    function onclickHandler(event) {
            // find the closest parent of the event target that
            // matches the selector
            const closest = event.target.closest(selector);
            if (closest) {
                // handle class event
                closest.classList.toggle("show");
            }
    }

    // add the handler to each element in the DOM that matches selector
    Array.from(_queryAll(selector)).forEach(function(element) {
        element.addEventListener('click', onclickHandler);
    });
}

function selectAndCopy(element) {
    element.select(); // select (highlight) text
    if (document.execCommand('copy') === true) { // copy to clipboard
        // copy to clipboard only works for short-running handlers,
        // so we check the return value and only display the tooltip
        // if copy to clipboard was successfull
        const tooltip = element.parentElement;
        tooltip.classList.add("show"); // show tooltip
        setTimeout(function(){ // hide tooltip after 5 seconds
            tooltip.classList.remove("show");
        }, 5000);
    }
}

function showQrCode(element, data) {
    if (! element) {
        console.warn("showQrCode: Element is not defined. Skipping.");
        return;
    }
    if (! window.QRCode) {
        console.warn("showQrCode: QRCode library not found. Skipping.");
        return;
    }

    /* legacy loading of the library, and we just checked that it exists */
    /* eslint-disable no-undef */
    const qrcode = new QRCode(element, {
        text: data,
        // width: 256,
        // height: 256,
        // colorDark : "#000000",
        // colorLight : "#ffffff",
        correctLevel : QRCode.CorrectLevel.M,
    });
    /* eslint-enable no-undef */

    setTimeout(function(){ // small delay until QR code is ready
        element.style.display = "block";
    }, 100);
}

// const linkList = new RecentItems('Link');
class RecentItems {
    constructor(itemName, wrapperElement) {
        this.itemName = itemName;
        this.wrapper = wrapperElement;
        this.list = [];
        if (! window.localStorage) {
            console.warn("RecentItems: Browser doesn't support localStorage. RecentItems disabled.");
            this.enabled = false;
            return;
        }
        this.enabled = true;
        this.localStorageName = "RecentItems_" + itemName;

        /* upgrade routine due to name change 'Link' to 'URL' -- 2020-11-10 */
        if (itemName == 'URL') {
            const oldName = 'RecentItems_Link';
            const oldData = localStorage.getItem(oldName);
            // if there is any old data and no new data
            if (oldData && ! localStorage.getItem(this.localStorageName)) {
                // store the old data under the new key
                localStorage.setItem(this.localStorageName, oldData);
                // delete the old key
                localStorage.removeItem(oldName);
            }
        }
    }

    // adds a new item to the list and returns the updated list
    append(link, data, timestamp) {
        // localSaveLink
        if (!this.enabled) {
            return [];
        }

        let item = {
            link: link,
            data: data,
            timestamp: timestamp,
        };

        let list = [];
        try {
            list = JSON.parse(localStorage.getItem(this.localStorageName) || '[]');
            list.push(item);
            localStorage.setItem(this.localStorageName, JSON.stringify(list));
        } catch(e) {
            console.error("RecentItems.append:", e);
        }

        this.list = list;
        return this.list;
    }

    // returns all the items in the current list
    all() {
        // localGetLinks
        if (!this.enabled) {
            return [];
        }

        let list = [];
        try {
            list = JSON.parse(localStorage.getItem(this.localStorageName) || '[]');
        } catch(e) {
            console.error("RecentItems.list:", e);
        }

        this.list = list;
        return this.list;
    }

    clear() {
        // clearLinkList
        if (!this.enabled) {
            return;
        }

        const tableBody = this.wrapper.querySelector("tbody");
        /* iteratively remove all HTML elements from table (this method also unregisters any event handlers) */
        while (tableBody.firstChild) {
            tableBody.removeChild(tableBody.firstChild);
        }

        /* clear localStorage */
        localStorage.removeItem(this.localStorageName);
    }

    fillHtml() {
        // partially populateLinkList

        /* retrieve list from localStorage */
        const list = this.all();
        if (list.length <= 0) {
            /* no items in the list, no need to do any work */
            return;
        }

        /* fill the wrapper with boilerplate for table */
        this.wrapper.innerHTML +=
        `<div class="table-height-limiter">
         <table id="${this.itemName}-list-table" class="pure-table pure-table-striped recent-items-table">
            <thead>
              <tr>
                <th>Link</th>
                <th>${this.itemName}</th>
                <th>Creation Date</th>
              </tr>
            </thead>
            <tbody>
            </tbody>
          </table>
          </div>`;
        const tableBody = this.wrapper.querySelector("tbody");

        /* iterate over the list in reverse-chronological order,
           so the most recent item is at the top */
        for (let i = list.length-1; i >= 0; i--) {
            const l = list[i];
            let row = "";
            try {
                // TODO: generalize this
                const link = l.link;
                const prettyLink = link.split('://')[1]; // strip leading https://
                const data = l.data;
                const ts = l.timestamp.split('T')[0] || ""; // just show the date
                if (this.itemName == 'URL') {
                    row = `
                    <td><a href="${link}" target="_blank">${prettyLink}</a></td>
                    <td><a href="${data}" target="_blank">${data}</a></td>
                    <td>${ts}</td>`;
                } else if (this.itemName == 'File') {
                    row = `
                    <td><a href="${link}" target="_blank">${prettyLink}</a></td>
                    <td>${data}</td>
                    <td>${ts}</td>`;
                }
            } catch(e) {
                console.warn("RecentItems.fillHtml: Error processing local link:", l, e);
            }

            /* if there were no errors (row != ""), append row to table */
            if (row.length > 0) {
                let child = document.createElement('tr');
                child.innerHTML = row;
                tableBody.append(child);
            }
        }

        /* if there are links in the local list, show the wrapper element to the user (by default hidden) */
        if (list.length > 0) {
            this.wrapper.style.display = "block";
        }
    }
}

function fileWidget() {
    let uploadFiles = [];

    const inputForm = _query('#file-input-form');
    const fileWrapper = _query("#file-wrapper");
    const inputSubmitButton = inputForm.submit;
    const fakeFileSelect = _query("#fake-file-input");
    const filePreviewWrapper = _query("#file-preview-wrapper");

    const outputForm = _query('#file-output-form');
    const outputField = outputForm.outputUrl;
    const sendEmailButton = outputForm.send;
    const qrCodeWrapper = outputForm.querySelector(".form-qr-code");
    outputField.addEventListener('focus', function(event) {
        selectAndCopy(event.target);
    });

    /* initialize the list of most recent links */
    const fileListWrapper = _query("#file-list-wrapper");
    const recentFiles = new RecentItems('File', fileListWrapper);
    recentFiles.fillHtml(); // populate the table
    /* register handler for deleting list */
    _query("#button-clear-file-list").addEventListener('click', (event) => {
        recentFiles.clear();
    });

    function displaySelectedFile(file) {
        if (file && 'name' in file && 'size' in file) {
            _query("#file-preview-details").innerHTML = `${file.name} - ${file.size/1000} kB`;
        }
    }

    function fileSelectHandler(e) {
        // cancel event and hover styling
        fileDragHoverHandler(e);
        const files = e.target.files || e.dataTransfer.files;
        // only supports uploading one file at a time
        uploadFiles = [files[0]];
        displaySelectedFile(files[0]);
    }

    function fileDragHoverHandler(e) {
        e.stopPropagation();
        e.preventDefault();
        (e.type == "dragover") ?
            fileWrapper.classList.add("highlighted") :
            fileWrapper.classList.remove("highlighted");
    }

    fileWrapper.addEventListener("dragenter", fileDragHoverHandler, false);
    fileWrapper.addEventListener("dragleave", fileDragHoverHandler, false);
    fileWrapper.addEventListener("dragover", fileDragHoverHandler, false);
    fileWrapper.addEventListener("drop", fileSelectHandler, false);

    // when someone clicks on the element, trigger the hidden file input
    // then the browser opens up a file picker dialog
    fakeFileSelect.addEventListener("change", fileSelectHandler, false);
    filePreviewWrapper.addEventListener("click", function() {
        fakeFileSelect.click();
    }, false);

    inputForm.addEventListener('submit', (event) => {
        // disable default action
        event.preventDefault();

        if (uploadFiles.length <= 0) {
            inputForm.querySelector("legend").innerHTML = "Please select a file before submitting:";
            console.error("no file specified");
            return;
        }

        // configure a request
        const xhr = new XMLHttpRequest();
        xhr.open('POST', '/file/');

        // prepare form data
        let data = new FormData();
        data.append("file", uploadFiles[0]);
        data.append("expire", inputForm.expire.value);

        // set up event handlers
        xhr.upload.addEventListener("progress", (e) => {
            let percent = Math.round(e.loaded / e.total * 100) || 100;
            inputSubmitButton.value = `${percent} %`;
        });
        xhr.addEventListener("load", () => {
            console.debug(xhr.responseText);
            if (xhr.readyState == 4) {
                switch (xhr.status) {
                case 200: {
                    const obj = JSON.parse(xhr.responseText);
                    const link = obj.link;
                    console.info(obj, link);
                    // show new URL and QR code
                    outputForm.style.display = "block";
                    outputField.value = link;
                    selectAndCopy(outputField);
                    showQrCode(qrCodeWrapper, link);
                    // hide original form (input and submit)
                    inputForm.style.display = "none";
                    // save in local storage
                    recentFiles.append(obj.link, obj.filename, obj.createTs);
                    break;
                }
                default:
                    console.error("XHR error:", xhr);
                    // change button style and text
                    inputSubmitButton.classList.add('pure-button-warning');
                    inputSubmitButton.value = "Try again";
                    inputForm.querySelector("legend").innerHTML = xhr.responseText;
                    break;
                }
            }
        });
        xhr.addEventListener("error", () => {
            console.error("XHR error:", xhr);
            // change button style and text
            inputSubmitButton.classList.add('pure-button-warning');
            inputSubmitButton.value = "Error!";
        });

        // send request
        xhr.send(data);
    });

    outputForm.addEventListener('submit', (event) => {
        // disable default action
        event.preventDefault();

        // indicate to the user that we are doing something
        // and block the button, just to be sure
        sendEmailButton.value = "Sending...";
        sendEmailButton.disabled = true;

        // configure a request
        const xhr = new XMLHttpRequest();
        xhr.open('POST', outputForm.outputUrl.value + '/email');

        // prepare form data
        let data = new FormData(outputForm);
        // data.append("to_email", outputForm.to_name);
        // data.append("from_name", outputForm.from_email);

        // set up event handlers
        xhr.addEventListener("load", () => {
            console.debug(xhr.responseText);
            if (xhr.readyState == 4) {
                switch (xhr.status) {
                case 200: {
                    sendEmailButton.value = "Done!";
                    sendEmailButton.classList.add('pure-button-ok');
                    break;
                }
                default:
                    console.error("XHR error:", xhr);
                    // change button style and text
                    sendEmailButton.classList.add('pure-button-warning');
                    sendEmailButton.value = "Error!";
                    break;
                }
            }
        });
        xhr.addEventListener("error", () => {
            console.error("XHR error:", xhr);
            // change button style and text
            sendEmailButton.classList.add('pure-button-warning');
            sendEmailButton.value = "Error!";
        });

        // send request
        xhr.send(data);
    });
}

function linkWidget() {
    const linkInputForm = _query('#link-input-form');
    const inputSubmitButton = linkInputForm.submit;

    const outputForm = _query('#link-output-form');
    const outputField = outputForm.outputUrl;
    const qrCodeWrapper = outputForm.querySelector(".form-qr-code");
    outputField.addEventListener('focus', function(event) {
        selectAndCopy(event.target);
    });

    /* initialize the list of most recent links */
    const linkListWrapper = _query("#link-list-wrapper");
    const recentLinks = new RecentItems('URL', linkListWrapper);
    recentLinks.fillHtml(); // populate the table
    /* register handler for deleting list */
    _query("#button-clear-link-list").addEventListener('click', (event) => {
        recentLinks.clear();
    });

    function submitHandler(event) {
        // disable default action
        event.preventDefault();

        // configure a request
        const xhr = new XMLHttpRequest();
        xhr.open('POST', '/link/');

        // prepare form data
        let data = new FormData(linkInputForm);

        // set up event handlers
        //xhr.addEventListener("progress", () => {});
        xhr.addEventListener("load", () => {
            console.debug(xhr.responseText);
            if (xhr.readyState == 4) {
                switch (xhr.status) {
                case 200: {
                    const obj = JSON.parse(xhr.responseText);
                    const link = obj.link;
                    // show new URL and QR code
                    outputForm.style.display = "block";
                    outputField.value = link;
                    selectAndCopy(outputField);
                    showQrCode(qrCodeWrapper, link);
                    // hide input form part (url and submit)
                    linkInputForm.style.display = "none";
                    // save in local storage
                    recentLinks.append(obj.link, obj.url, obj.createTs);
                    break;
                }
                default:
                    console.error("XHR error:", xhr);
                    // change button style and text
                    inputSubmitButton.classList.add('pure-button-warning');
                    inputSubmitButton.value = "Try again";
                    linkInputForm.querySelector("legend").innerHTML = xhr.responseText;
                    break;
                }
            }
        });
        xhr.addEventListener("error", () => {
            console.error("XHR error:", xhr);
            // change button style and text
            inputSubmitButton.classList.add('pure-button-warning');
            inputSubmitButton.value = "Error!";
        });

        // send request
        xhr.send(data);
    }

    linkInputForm.addEventListener('submit', submitHandler);
}
