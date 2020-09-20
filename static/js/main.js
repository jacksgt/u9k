function selectAndCopy(element) {
    element.select(); // select (highlight) text
    document.execCommand('copy'); // copy to clipboard
    const tooltip = element.parentElement;
    tooltip.classList.toggle("show"); // show tooltip
    setTimeout(function(){ // hide tooltip after 5 seconds
        tooltip.classList.toggle("show");
    }, 5000);
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
	    // width: 128,
	    // height: 128,
	    // colorDark : "#000000",
	    // colorLight : "#ffffff",
	    // correctLevel : QRCode.CorrectLevel.H
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
        let child = document.createElement('tr');
        // TODO: sanitize the treatment of these JSON fields
        child.innerHTML = `
              <td><a href="${l.link}" target="_blank">${l.link.split('://')[1]}</a></td>
              <td><a href="${l.url}" target="_blank">${l.url}</a></td>
              <td>${l.createTs.split('T')[0]}</td>
          `;
        tableBody.append(child);
    }

    /* if there are links in the local list, show the wrapper element to the user (by default hidden) */
    if (linkList.length > 0) {
        tableWrapper.style.display = "block";
    }
}

function registerLinkFormHandler() {
    const linkForm = document.querySelector('#link-form');
    linkForm.addEventListener('submit', (event) => {
        // disable default action
        event.preventDefault();

        // configure a request
        const xhr = new XMLHttpRequest();
        xhr.open('POST', '/link/');

        // prepare form data
        let data = new FormData(linkForm);

        // set headers
        // xhr.setRequestHeader('Content-Type', 'multipart/form-data');

        // set up event handlers
        xhr.addEventListener("progress", () => {});
        xhr.addEventListener("load", () => {
            console.log(xhr.responseText);
            if (xhr.readyState == 4) {
                switch (xhr.status) {
                case 200:
                    const obj = JSON.parse(xhr.responseText);
                    const link = obj.link;
                    // change button style and text
                    linkForm.querySelector("legend").innerHTML = "Your link is now available under the following URL:"
                    var button = linkForm.submit;
                    button.classList.add('pure-button-ok');
                    button.value = "Done!";
                    // update with new URL
                    const field = linkForm.url;
                    field.value = link;
                    field.readOnly = true;
                    selectAndCopy(field);
                    // hide optional form part (input and submit)
                    linkForm.querySelector(".optional-form-part").style.display = "none";
                    showQrCode(linkForm.querySelector(".form-qr-code"), link)
                    // save in local storage
                    localSaveLink(obj);
                    break;

                default:
                    console.log("ERROR:", xhr);
                    // change button style and text
                    var button = linkForm.submit;
                    button.classList.add('pure-button-warning');
                    button.value = "Try again ...";
                    linkForm.querySelector("legend").innerHTML = xhr.responseText;
                    break;
                }
            }
        });
        xhr.addEventListener("error", () => {
            console.log("ERROR:", xhr);
            // change button style and text
            const button = linkForm.submit;
            button.classList.add('pure-button-orange');
            button.value = "Oops, there was an error!";
        });

        // send request
        xhr.send(data);
    });
}

/* register all event handlers */
window.addEventListener('load', (event) => {
    registerLinkFormHandler();
    populateLinkList();
    fileForm();
});
