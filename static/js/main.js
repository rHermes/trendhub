function langTitleClick(lang) {
  const b = document.querySelector("#lang-" + lang + " .trending-item-list");
  b.classList.toggle("cloaked");
}

function isScrolledIntoView(el) {
    var rect = el.getBoundingClientRect();
    var elemTop = rect.top;
    var elemBottom = rect.bottom;

    // Only completely visible elements return true:
    //var isVisible = (elemTop >= 0) && (elemBottom <= window.innerHeight);
    // Partially visible elements return true:
    var isVisible = elemTop < window.innerHeight && elemBottom >= 0;
    return isVisible;
}



(function() {
  'use strict';

  function sidebarHighlightFunc() {
    LANGUAGES.forEach((x) => {
      let a = document.getElementById("lang-" + x);
      let b = document.getElementById("navbar-lang-" + x);
      if (isScrolledIntoView(a)) {
        b.classList.add("navbar-lang-active")
        if (!isScrolledIntoView(b)) {
//           b.scrollIntoView({ behavior: 'auto', block: 'nearest', inline: 'start' })
          //b.scrollIntoView(true);
        }
      } else {
        b.classList.remove("navbar-lang-active")
      }
    })
  };
  sidebarHighlightFunc();

  // Initial setup
  document.addEventListener("scroll", sidebarHighlightFunc);
})();
