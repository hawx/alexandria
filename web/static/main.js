/* global fetch, EventSource, FormData */

function h (tag, attrs = {}, children = []) {
  const el = document.createElement(tag)

  for (const attr in attrs) {
    if (attr.startsWith('on')) {
      el[attr] = e => attrs[attr](el, e)
    } else {
      el.setAttribute(attr, attrs[attr])
    }
  }

  for (const child of children) {
    if (typeof child === 'string') {
      el.appendChild(document.createTextNode(child))
    } else {
      el.appendChild(child)
    }
  }

  return el
}

function editionsTmpl (data) {
  return data.editions.map(edition => h('a', { href: edition.links.self.href }, [edition.name]))
}

class Row {
  constructor (parent, id) {
    this.parent = parent
    this.id = id
  }

  delete () {
    return fetch(`/books/${this.id}`, { method: 'DELETE' })
  }

  setTitle (title) {
    return fetch(`/books/${this.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ title })
    })
  }

  setAuthor (author) {
    return fetch(`/books/${this.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ author })
    })
  }

  render (data, next) {
    const output = h('tr', { id: data.id }, [
      h('td', {}, [
        h('input', {
          class: 'title',
          value: data.title,
          onblur: me => this.setTitle(me.value),
          onkeypress: (me, e) => {
            if (e.keyCode === 13) me.blur()
          }
        })
      ]),
      h('td', {}, [
        h('input', {
          class: 'author',
          value: data.author,
          onblur: me => this.setAuthor(me.value),
          onkeypress: (me, e) => {
            if (e.keyCode === 13) me.blur()
          }
        })
      ]),
      h('td', {}, [data.added]),
      h('td', { class: 'editions' }, editionsTmpl(data))
      // h('td', { class: 'delete' }, [
      //   h('a', {
      //     href: '#',
      //     class: 'delete',
      //     onclick: () => this.delete(),
      //   }, [
      //     'delete',
      //   ]),
      // ])
    ])

    this.parent.querySelector('tbody').appendChild(output)

    next && next()
  }

  remove () {
    const parent = this.parent.querySelector('tbody')
    const el = document.getElementById(this.id)

    parent.removeChild(el)
  }

  update (data, next) {
    const el = document.getElementById(this.id)

    el.querySelector('.title').value = data.title
    el.querySelector('.author').value = data.author

    const editions = el.querySelector('.editions')

    for (const child of editions.children) {
      editions.removeChild(child)
    }
    for (const child of editionsTmpl(data)) {
      editions.appendChild(child)
    }
  }

  add (data, next) {
    this.render(data, next)
  }
}

class Rows {
  constructor (parent) {
    this.parent = parent
    this._rows = {}
  }

  render (next) {
    return fetch('/books')
      .then(resp => resp.json())
      .then(data => {
        for (const book of data.books) {
          this.add(book.id, book)
        }

        next && next()
      })
  }

  add (id, data) {
    var row = new Row(this.parent, id)
    this._rows[id] = row
    row.add(data)
  }

  addTemp (name) {
    this.temp = h('tr', { class: 'temp' }, [
      h('td', {}, ['Uploading']),
      h('td', {}, [name]),
      h('td'),
      h('td')
    ])

    this.parent.appendChild(this.temp)
  }

  removeTemp () {
    this.parent.removeChild(this.temp)
  }

  update (id, data) {
    this._rows[id].update(data)
  }

  remove (id) {
    this._rows[id].remove()
  }
}

function upload (rows, files) {
  const formData = new FormData()
  for (const file of files) {
    rows.addTemp(file.name)
    formData.append('file', file)
  }

  return fetch('/upload', {
    method: 'POST',
    body: formData
  })
}

function cancel (fn) {
  return function (event) {
    fn(event)

    if (event.preventDefault) {
      event.preventDefault()
    }
    return false
  }
}

function sortable (table, initial) {
  const ths = table.querySelectorAll('th')
  let sortIndex = initial; let sortDirection = -1

  function sortOn (heading, index) {
    return () => {
      if (sortIndex === index) {
        sortDirection *= -1

        if (sortDirection === 1) {
          heading.classList.add('descending')
          heading.classList.remove('ascending')
        } else {
          heading.classList.add('ascending')
          heading.classList.remove('descending')
        }
      } else {
        ths[sortIndex].classList.remove('ascending')
        ths[sortIndex].classList.remove('descending')

        sortIndex = index
        sortDirection = 1

        heading.classList.add('descending')
      }

      resort()
    }
  }

  function resort () {
    const tbody = table.querySelector('tbody')
    const trs = tbody.querySelectorAll('tr')
    const mapped = []

    for (const tr of trs) {
      const td = tr.querySelectorAll('td')[sortIndex]
      const text = td.firstChild.value || td.firstChild.textContent

      mapped.push([text, tr])
      tbody.removeChild(tr)
    }

    mapped.sort((a, b) => {
      return a[0] < b[0] ? -sortDirection : sortDirection
    })

    for (const [, row] of mapped) {
      tbody.appendChild(row)
    }
  }

  for (let i = 0; i < ths.length; i++) {
    ths[i].onclick = sortOn(ths[i], i)
  }

  return [sortOn(ths[sortIndex], sortIndex), resort]
}

function filterable (table, filter, next) {
  let hidden = []

  filter.onkeyup = () => {
    const tbody = table.querySelector('tbody')
    const trs = tbody.querySelectorAll('tr')
    const needle = filter.value

    if (needle === '') {
      for (const tr of hidden) {
        tbody.appendChild(tr)
      }
      hidden = []
      next()
      return
    }

    for (const tr of trs) {
      const tds = tr.querySelectorAll('td')
      let match = false

      for (const td of tds) {
        const text = td.firstChild.value || td.firstChild.textContent

        if (text.toLowerCase().includes(needle.toLowerCase())) {
          match = true
          break
        }
      }

      if (!match) {
        hidden.push(tr)
        tbody.removeChild(tr)
      }
    }

    const newHidden = []
    for (const tr of hidden) {
      const tds = tr.querySelectorAll('td')
      let match = false

      for (const td of tds) {
        const text = td.firstChild.value || td.firstChild.textContent

        if (text.toLowerCase().includes(needle.toLowerCase())) {
          match = true
          break
        }
      }

      if (!match) {
        newHidden.push(tr)
      } else {
        tbody.appendChild(tr)
      }
    }

    hidden = newHidden
    next()
  }
}

const body = document.body
const table = document.querySelector('table')
const rows = new Rows(table)

const [initial, resort] = sortable(table, 1)
filterable(table, document.querySelector('#filter'), resort)
rows.render(initial)

document.ondragenter = cancel(() => { body.className = 'drag' })
document.ondragover = cancel(() => { body.className = 'drag' })
document.ondragleave = cancel(() => { body.className = '' })

document.ondrop = cancel(ev => {
  body.className = ''
  upload(rows, ev.dataTransfer.files)
})

const es = new EventSource('/events')

es.addEventListener('add', e => {
  rows.removeTemp()
  const obj = JSON.parse(e.data)
  rows.add(obj.id, obj)
}, false)

es.addEventListener('update', e => {
  const obj = JSON.parse(e.data)
  rows.update(obj.id, obj)
}, false)

es.addEventListener('delete', e => {
  const obj = JSON.parse(e.data)
  rows.remove(obj.id)
}, false)
