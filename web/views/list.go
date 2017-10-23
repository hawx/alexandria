package views

const list = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>alexandria</title>
    <link rel="stylesheet" href="/assets/styles.css" />
  </head>

  <body>
    <h1>alexandria</h1>

    {{if .}}
      <div id="cover">
        <a href="/sign-in">Sign-in</a>
      </div>
    {{end}}

    <input id="filter" name="filter" type="text" placeholder="Search" />

    <table>
      <thead>
        <tr>
          <th>Title</th>
          <th>Author</th>
          <th>Added</th>
          <th>Editions</th>
        </tr>
      </thead>
      <tbody></tbody>
    </table>

    <script src="//code.jquery.com/jquery-2.1.1.min.js"></script>
    <script src="/assets/mustache.js"></script>
    <script src="/assets/tablefilter.js"></script>
    <script src="/assets/tablesorter.js"></script>
    <script src="/assets/main.js"></script>
  </body>
</html>`
