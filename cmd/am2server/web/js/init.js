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

    var elemsCollapsible = content.querySelectorAll('.collapsible');
    var instancesCollapsible = M.Collapsible.init(elemsCollapsible, options);

    $('.materialert .close-alert').click(function (){
        $(this).parent().hide('slow');
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