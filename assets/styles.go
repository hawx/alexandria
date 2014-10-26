package assets

const Styles = `body {
    font: 16px/1.3em Helvetica;
    width: 100%;
    margin: 0;
    padding: 0;
    transition: opacity .1s;
    overflow: hidden;
}

body.drag {
    opacity: .3;
}

h1 {
    margin: 2rem;
    font: italic 28px/1.3em Georgia;
}

#filter {
    position: absolute;
    top: 2em;
    right: 2em;
    padding: 4px;
    width: 280px;
}

table {
    border-collapse: collapse;
    width: 100%;
    margin: 2em 0 0;
}

thead {
    border-bottom: 1px solid;
}

th:first-child, td:first-child {
    padding-left: 2em;
}

th:last-child, td:last-child {
    padding-right: 2em;
}

th, td {
    padding: .5em 0;
}

tbody tr:nth-child(odd) {
    background: #efefef;
}

table th {
    text-align: left;
}

th::after {
    position: absolute;
    margin-left: 10px;
    font-size: 13px;
    margin-top: -1px;
    color: grey;
}

th.ascending::after {
    content: '↑';
}

th.descending::after {
    content: '↓';
}

td a {
    margin-right: 5px;
}

.author, .title {
    border: none;
    background: none;
    color: black;
    font: 16px/1.3em Helvetica;
    border-bottom: 1px solid transparent;
    width: 90%;
}

.author:hover, .title:hover {
    border-bottom: 1px dotted silver;
}

.author:focus, .title:focus {
    border-bottom: 1px dotted black;
}

#cover {
    top: 0;
    left: 0;
    z-index: 1000;
    position: absolute;
    height: 100%;
    width: 100%;
    background: rgba(0, 255, 255, .7);
    display: block
    padding: 0;
    margin: 0;
}

#cover a {
    position: relative;
    display: block;
    left: 50%;
    top: 50%;
    text-align: center;
    width: 100px;
    margin-left: -50px;
    height: 50px;
    line-height: 50px;
    margin-top: -25px;
    font-size: 16px;
    font-weight: bold;
    border: 1px solid;
}`
