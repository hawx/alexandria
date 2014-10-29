package assets

const MainJs = `var http = (function() {
  var request = function(method, url, success, error, data) {
    var sendingData = (data !== void 0);

    var r = new XMLHttpRequest();
    r.open(method, url);
    r.setRequestHeader('HTTP_X_CSRF_TOKEN', window.CSRF_TOKEN);
    r.setRequestHeader('Accept', 'application/json');
    if (sendingData) {
      r.setRequestHeader('Content-Type', 'application/json');
    }

    r.onload = function() {
      if (r.status >= 200 && r.status < 300) {
        success && success(JSON.parse(r.response));
      } else {
        error && error(Error(r.statusText));
      }
    };

    r.onerror = function() {
      error && error(Error("Network error"));
    };

    if (sendingData) {
      r.send(JSON.stringify(data));
    } else {
      r.send();
    }
  };

  return {
    get: function(obj) {
      request('GET', obj.url, obj.success, obj.error);
    },
    delete: function(obj) {
      request('DELETE', obj.url, obj.success, obj.error);
    },
    post: function(obj) {
      var r = new XMLHttpRequest();
      r.open('POST', obj.url);
      r.onload = function(e) {
        if (r.status != 204) {
          obj.error && obj.error(Error(r.statusText));
        }
      };

      r.send(obj.formData);
    },
    patch: function(obj) {
      request('PATCH', obj.url, obj.success, obj.error, obj.data);
    }
  };
})();


var Row = function(parent, id) {
  var _id = id,
      tmpl = "<tr id=\"{{id}}\">" +
        "<td><input type=\"text\" class=\"title\" value=\"{{title}}\" /></td>" +
        "<td><input type=\"text\" class=\"author\" value=\"{{author}}\" /></td>" +
        "<td>{{added}}</td>" +
        "<td class=\"editions\">" +
        "  {{#editions}}" +
        "    <a href=\"{{links.self.href}}\">{{name}}</a>" +
        "   {{/editions}}" +
        "</td>" +
        "<!-- <td>" +
        "  <a href=\"#\" class=\"delete\">delete</a>" +
        "</td> -->" +
        "</tr>",
      editionsTmpl = "{{#editions}}" +
        "  <a href=\"{{links.self.href}}\">{{name}}</a>" +
        "{{/editions}}";

  var deleteEvent = function() {
    http.delete({url: '/books/' + _id});
  };

  var setTitleEvent = function() {
    http.patch({
      url: '/books/' + _id,
      data: {title: this.val()}
    });
  };

  var setAuthorEvent = function() {
    http.patch({
      url: '/books/' + _id,
      data: {author: this.val()}
    });
  };

  var render = function(output, next) {
    parent.append(output);

    var del = parent.find('#' + _id + ' .delete');
    del.click(deleteEvent);

    var title = parent.find('#' + _id + ' .title');
    title.blur(setTitleEvent.bind(title));
    title.keypress(function(e) {
      if (e.keyCode == 13) {
        title.blur();
      }
    });

    var author = parent.find('#' + _id + ' .author');
    author.blur(setAuthorEvent.bind(author));
    author.keypress(function(e) {
      if (e.keyCode == 13) {
        author.blur();
      }
    });

    next && next();
  };

  this.remove = function() {
    parent.find('#' + _id).remove();
  };

  this.update = function(data, next) {
    var el = parent.find('#' + _id);
    el.find('.title').val(data.title);
    el.find('.author').val(data.author);

    var editionsText = Mustache.render(editionsTmpl, data);
    el.find('.editions').html(editionsText);
  };

  this.add = function(data, next) {
    var output = Mustache.render(tmpl, data);
    render(output, next);
  };
};

var Rows = function(parent) {
  this.render = function(next) {
    http.get({
      url: '/books',
      success: function(data) {
        for (var i = 0; i < data.books.length; i++) {
          var row = new Row(parent, data.books[i].id);
          _rows[data.books[i].id] = row;
          row.add(data.books[i]);
        }

        next && next();
      }
    });
  };

  var _rows = {};

  this.add = function(id, data) {
    var row = new Row(parent, id);
    _rows[id] = row;
    row.add(data);
  };

  this.addTemp = function(name) {
    var t = "<tr class=\"temp\">" +
          "  <td>Uploading</td>" +
          "  <td>" + name + "</td>" +
          "  <td></td>" +
          "  <td></td>" +
          "</tr>";

    parent.append(t);
  };

  this.removeTemp = function() {
    parent.find(".temp")[0].remove();
  };

  this.update = function(id, data) {
    _rows[id].update(data);
  };

  this.remove = function(id) {
    _rows[id].remove();
  };
};

function gotAssertion(assertion) {
  // got an assertion, now send it up to the server for verification
  if (assertion !== null) {
    $.ajax({
      type: 'POST',
      url: '/sign-in',
      data: { assertion: assertion },
      success: function(res, status, xhr) {
        window.location.reload();
      },
      error: function(xhr, status, res) {
        alert("login failure" + res);
      }
    });
  }
};

function connect(rows) {
  var es = new EventSource('/events');

  es.addEventListener("add", function(e) {
    rows.removeTemp();
    var obj = JSON.parse(e.data);
    rows.add(obj.id, obj);
  }, false);

  es.addEventListener("update", function(e) {
    var obj = JSON.parse(e.data);
    rows.update(obj.id, obj);
  }, false);

  es.addEventListener("delete", function(e) {
    var obj = JSON.parse(e.data);
    rows.remove(obj.id);
  }, false);
}

function upload(rows, files) {
  var formData = new FormData();
  for (var i = 0; i < files.length; i++) {
    rows.addTemp(files[i].name);
    formData.append('file', files[i]);
  }

  http.post({
    url: '/upload',
    formData: formData
  });
}

function cancel(fn) {
  return function(event) {
    fn(event);

    if (event.preventDefault) {
      event.preventDefault();
    }
    return false;
  };
}

$(function($) {
  $('#browserid').click(function() {
    navigator.id.get(gotAssertion);
  });

  var body = document.body;
  var table = $('table');
  var rows = new Rows(table);

  rows.render(function() {
    table
      .trigger("update")
      .trigger("appendCache")
      .trigger('sortOn', [[1, 0]]);
  });

  document.ondragenter = cancel(function(ev) {
    body.className = 'drag';
  });

  document.ondragover = cancel(function(ev) {
    body.className = 'drag';
  });

  document.ondragleave = cancel(function(ev) {
    body.className = '';
  });

  document.ondrop = cancel(function(ev) {
    body.className = '';
    upload(rows, ev.dataTransfer.files);
  });

  connect(rows);

  $('#filter').keyup(function() {
    $.tableFilter(table, this.value);
  });

  table.tablesorter({
    sortList: [[2,1]],
    cssAsc: 'ascending',
    cssDesc: 'descending',
    textExtraction: function(node) {
      if (node.childNodes[0].nodeName === "#text") {
        return node.childNodes[0].textContent;
      }

      return node.childNodes[0].value;
    }
  });
});`
