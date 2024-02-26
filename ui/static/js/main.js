// var navLinks = document.querySelectorAll("nav a");
// for (var i = 0; i < navLinks.length; i++) {
// 	var link = navLinks[i]
// 	if (link.getAttribute('href') == window.location.pathname) {
// 		link.classList.add("live");
// 		break;
// 	}
// }

const regInput = document.getElementById("reg_input");
regInput.value = "";

const regSubmit = document.getElementById("reg_submit");

regSubmit.addEventListener("click", function(event) {
    if (regInput.value != "") {
        window.location.href = `/aircraft?reg=${regInput.value}`
    }
})

regInput.addEventListener("keypress", function(event) {
    if (event.key == "Enter") {
        regSubmit.click();
    }
})

