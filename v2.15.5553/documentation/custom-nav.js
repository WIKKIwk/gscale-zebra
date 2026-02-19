document.addEventListener("DOMContentLoaded", function () {
    var navbar = document.querySelector("#navbar-top-firstrow");
    if (navbar) {
        var newItem = document.createElement("li");
        newItem.innerHTML = '<a href="allclasses-index.html">ALL CLASSES</a>';
        navbar.appendChild(newItem);
    }
});