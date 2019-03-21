(function($) {
  $.tableFilter = function(jq, phrase, ifHidden, tdElem) {
    if (!tdElem) tdElem = "td";
    var new_hidden = false;
    if (this.last_phrase === phrase) return false;

    var phrase_length = phrase.length;
    var words = phrase.toLowerCase().split(" ");

    // these function pointers may change
    var matches = function(elem) { elem.show(); };
    var noMatch = function(elem) { elem.hide(); new_hidden = true; };
    var getText = function(elem) {
      var extract = function(node) {
        if (node.childNodes.length != 1) { return ""; }

        return node.childNodes[0].nodeName === "#text"
          ? node.childNodes[0].textContent
          : node.childNodes[0].value;
      };

      var s = "";
      for (var i = 0; i < elem[0].childNodes.length; i++) {
        s += extract(elem[0].childNodes[i]) + " ";
      };
      return s;
    };

    // if added one letter to last time,
    // just check newest word and only need to hide
    if((words.size > 1) && (phrase.substr(0, phrase_length - 1) === this.last_phrase)) {
      if (phrase[-1] === " ") { this.last_phrase = phrase; return false; }

      var words = words[-1]; // just search for the newest word

      // only hide visible rows
      matches = function(elem) {};
      var elems = jq.find("tbody:first > tr:visible");
    } else {
      new_hidden = true;
      var elems = jq.find("tbody:first > tr");
    }

    elems.each(function(){
      var elem = $(this);
      $.tableFilter.has_words(getText(elem), words, false)
        ? matches(elem)
        : noMatch(elem);
    });

    $.tableFilter.last_phrase = phrase;
    if( ifHidden && new_hidden ) ifHidden();
    return jq;
  };

  // caching for speedup
  $.tableFilter.last_phrase = "";

  // not jQuery dependent
  // "" [""] -> Boolean
  // "" [""] Boolean -> Boolean
  $.tableFilter.has_words = function(str, words, caseSensitive) {
    var text = caseSensitive ? str : str.toLowerCase();
    for (var i = 0; i < words.length; i++) {
      if (text.indexOf(words[i]) === -1) return false;
    }
    return true;
  };
}) (jQuery);
