function nextStep() {
    const step1 = document.getElementById("step1");
    const step2 = document.getElementById("step2");
    const step3 = document.getElementById("step3");

    if (!step1.classList.contains("hidden")) {
        step1.classList.add("hidden");
        step2.classList.remove("hidden");
    } else if (!step2.classList.contains("hidden")) {
        step2.classList.add("hidden");
        step3.classList.remove("hidden");
    }
}