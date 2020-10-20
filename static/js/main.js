function _query(query) {
    return document.querySelector(query);
}

function _queryAll(query) {
    return document.querySelectorAll(query);
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
        console.log("showQrCode: Element is not defined. Skipping.");
        return;
    }
    if (! window.QRCode) {
        console.log("showQrCode: QRCode library not found. Skipping.");
        return;
    }

    const qrcode = new QRCode(element, {
	    text: data,
	    // width: 256,
	    // height: 256,
	    // colorDark : "#000000",
	    // colorLight : "#ffffff",
	    correctLevel : QRCode.CorrectLevel.M,
    });

    setTimeout(function(){ // small delay until QR code is ready
        element.style.display = "block";
    }, 100);
}

function localSaveLink(link) {
    if (! window.localStorage) {
        console.log("localSaveLink: Browser doesn't support localStorage. Skipping.");
        return []
    }

    let linkList;
    try {
        linkList = JSON.parse(localStorage.getItem("linkList") || '[]');
        linkList.push(link);
        localStorage.setItem("linkList", JSON.stringify(linkList));
    } catch(e) {
        console.log("localSaveLink:", e);
    }

    return linkList;
}

function localGetLinks() {
    if (! window.localStorage) {
        console.log("localGetLinks: Browser doesn't support localStorage. Skipping.");
        return []
    }

    let linkList = [];
    try {
        linkList = JSON.parse(localStorage.getItem("linkList") || '[]');
    } catch(e) {
        console.log("localGetLinks:", e);
    }
    return linkList;
}

function clearLinkList() {
    const tableBody = document.querySelector('#link-list-table tbody');
    /* iteratively remove all HTML elements from table (this method also unregisters any event handlers) */
    while (tableBody.firstChild) {
        tableBody.removeChild(tableBody.firstChild);
    }
    /* clear localStorage entry */
    if (window.localStorage) {
        localStorage.removeItem('linkList');
    }
}

function populateLinkList() {
    /* register handler for deleting list */
    document.querySelector("#button-clear-link-list").addEventListener('click', (event) => {
        clearLinkList();
    })

    const tableWrapper = document.querySelector("#link-list-wrapper");
    const tableBody = tableWrapper.querySelector("tbody");

    /* retrieve links stored in localStorage */
    const linkList = localGetLinks();

    /* iterate over the list in reverse-chronological order,
       so the most recent item is at the top */
    for (let i = linkList.length-1; i >= 0; i--) {
        const l = linkList[i];
        let text = "";
        try {
            const link = l.link;
            const prettyLink = link.split('://')[1];
            const url = l.url;
            const ts = l.createTs.split('T')[0] || "";
            text = `
              <td><a href="${link}" target="_blank">${prettyLink}</a></td>
              <td><a href="${url}" target="_blank">${url}</a></td>
              <td>${ts}</td>
          `;
        } catch(e) {
            console.log("Error processing local link:", l, e);
        }
        if (text != "") {
            let child = document.createElement('tr');
            child.innerHTML = text;
            tableBody.append(child);
        }

    }

    /* if there are links in the local list, show the wrapper element to the user (by default hidden) */
    if (linkList.length > 0) {
        tableWrapper.style.display = "block";
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
            inputForm.querySelector("legend").innerHTML = "Please select a file before submitting:"
            console.log("ERROR: no file specified");
            return
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
            console.log(xhr.responseText);
            if (xhr.readyState == 4) {
                switch (xhr.status) {
                case 200:
                    const obj = JSON.parse(xhr.responseText);
                    const link = obj.link;
                    console.log(obj, link);
                    // show new URL and QR code
                    outputForm.style.display = "block";
                    outputField.value = link;
                    selectAndCopy(outputField);
                    showQrCode(qrCodeWrapper, link)
                    // hide original form (input and submit)
                    inputForm.style.display = "none";
                    // // save in local storage
                    // localSaveLink(obj);
                    // TODO: implement recent file list
                    break;

                default:
                    console.log("ERROR:", xhr);
                    // change button style and text
                    inputSubmitButton.classList.add('pure-button-warning');
                    inputSubmitButton.value = "Try again";
                    inputForm.querySelector("legend").innerHTML = xhr.responseText;
                    break;
                }
            }
        });
        xhr.addEventListener("error", () => {
            console.log("ERROR:", xhr);
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

        // configure a request
        const xhr = new XMLHttpRequest();
        xhr.open('POST', outputForm.outputUrl.value + '/email');

        // prepare form data
        let data = new FormData(outputForm);
        // data.append("to_email", outputForm.to_name);
        // data.append("from_name", outputForm.from_email);

        // set up event handlers
        xhr.addEventListener("load", () => {
            console.log(xhr.responseText);
            if (xhr.readyState == 4) {
                switch (xhr.status) {
                case 200:
                    sendEmailButton.value = "Done!";
                    sendEmailButton.classList.add('pure-button-ok');
                    sendEmailButton.disabled = true; // to make sure user does not click multiple times
                    break;
                default:
                    console.log("ERROR:", xhr);
                    // change button style and text
                    sendEmailButton.classList.add('pure-button-warning');
                    sendEmailButton.value = "Error!";
                    break;
                }
            }
        });
        xhr.addEventListener("error", () => {
            console.log("ERROR:", xhr);
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
            console.log(xhr.responseText);
            if (xhr.readyState == 4) {
                switch (xhr.status) {
                case 200:
                    const obj = JSON.parse(xhr.responseText);
                    const link = obj.link;
                    // show new URL and QR code
                    outputForm.style.display = "block";
                    outputField.value = link;
                    selectAndCopy(outputField);
                    showQrCode(qrCodeWrapper, link)
                    // hide input form part (url and submit)
                    linkInputForm.style.display = "none";
                    // save in local storage
                    localSaveLink(obj);
                    break;

                default:
                    console.log("ERROR:", xhr);
                    // change button style and text
                    inputSubmitButton.classList.add('pure-button-warning');
                    inputSubmitButton.value = "Try again";
                    linkInputForm.querySelector("legend").innerHTML = xhr.responseText;
                    break;
                }
            }
        });
        xhr.addEventListener("error", () => {
            console.log("ERROR:", xhr);
            // change button style and text
            inputSubmitButton.classList.add('pure-button-warning');
            inputSubmitButton.value = "Error!";
        });

        // send request
        xhr.send(data);
    }

    linkInputForm.addEventListener('submit', submitHandler);
}

/* register all event handlers */
window.addEventListener('load', (event) => {
    linkWidget();
    populateLinkList();
    fileWidget();
});
