const regInput = document.getElementById("reg_input");
regInput.value = "";

const regSubmit = document.getElementById("reg_submit");

regSubmit.addEventListener("click", function (_) {
    if (regInput.value != "") {
        window.location.href = `/aircraft?reg=${regInput.value}`;
    }
});

regInput.addEventListener("keypress", function (event) {
    if (event.key == "Enter") {
        regSubmit.click();
    }
});
