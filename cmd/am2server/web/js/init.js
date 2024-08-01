loadComponents(document);


htmx.onLoad(function (content) {
    loadComponents(content);
});

function loadComponents(content) {
    var elems = content.querySelectorAll('select');
    var options = {};
    var instances = M.FormSelect.init(elems, options);


    var elemsSidenav = content.querySelectorAll('.sidenav');
    var instancesSidenav = M.Sidenav.init(elemsSidenav, options);

    var elemsCollapsible = document.querySelectorAll('.collapsible');
    var instancesCollapsible = M.Collapsible.init(elemsCollapsible, options);

    $('.materialert .close-alert').click(function () {
        $(this).parent().hide('slow');
    });

    $(document).ready(function () {
        $('.tabs').tabs();
    });
}

function replacePathParams(event) {
    let pathWithParameters = event.detail.path.replace(/{([A-Za-z0-9_]+)}/g, function (_match, parameterName) {
        let parameterValue = event.detail.parameters[parameterName]
        delete event.detail.parameters[parameterName]
        return parameterValue
    })
    event.detail.path = pathWithParameters
}

function showMessage(msg) {
    var msgIcon = 'check_circle'
    switch (msg.type) {
        case 'error':
            msgIcon = 'error_outline'
            break
        case 'warning':
            msgIcon = 'warning'
            break
        case 'info':
            msgIcon = 'info_outline'
            break
        case 'success':
            msgIcon = 'check'
            break
    }
    const messageDiv =
        `<div class="materialert ` + msg.type + `">
            <div class="material-icons">` + msgIcon + `</div>
            <span>` + msg.text + `</span>
            <button type="button" class="close-alert">×</button>
        </div>`
    var messages = htmx.find('#messages');
    console.log('messages', messages);
    messages.innerHTML = messageDiv;
    $('.materialert .close-alert').click(function () {
        $(this).parent().hide('slow');
    });
}

htmx.on('htmx:responseError', function (evt) {
    try {
        const msg = JSON.parse(evt.detail.xhr.response)
        showMessage(msg)
    } catch (e) {
        const msg = {
            type: 'error',
            text: evt.detail.xhr.response
        }
        showMessage(msg)
    }
});

htmx.on('htmx:sendError', function () {
    const msg = {
        type: 'warning',
        text: 'Server unavailable. Try again in a few minutes.'
    }
    showMessage(msg)
});

function onSignIn(googleUser) {
    var profile = googleUser.getBasicProfile();
    console.log('ID: ' + profile.getId()); // Do not send to your backend! Use an ID token instead.
    console.log('Name: ' + profile.getName());
    console.log('Image URL: ' + profile.getImageUrl());
    console.log('Email: ' + profile.getEmail()); // This is null if the 'email' scope is not present.
}

function signOut() {
    var auth2 = gapi.auth2.getAuthInstance();
    auth2.signOut().then(function () {
        console.log('User signed out.');
    });
}