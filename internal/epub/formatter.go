package epub

const DefaultCSS = `
body {
	font-family: Georgia, serif;
	line-height: 1.6;
	color: #333;
}
h1 {
	color: #2c3e50;
	text-align: center;
	border-bottom: 2px solid #3498db;
	padding-bottom: 10px;
}
p {
	margin-bottom: 1em;
	text-align: justify;
}
.red { color: #e74c3c; }
.blue { color: #3498db; }
.green { color: #27ae60; }
.purple { color: #9b59b6; }
.orange { color: #e67e22; }
.yellow { color: #f1c40f; }
.brown { color: #8b4513; }
.pink { color: #e91e63; }
.cyan { color: #1abc9c; }
.gray { color: #7f8c8d; }
.gold { color: #ffd700; }
.silver { color: #c0c0c0; }
.crimson { color: #dc143c; }
.maroon { color: #800000; }
.navy { color: #000080; }
.teal { color: #008080; }
`

type Formatter struct {
	css string
}

func NewFormatter() *Formatter {
	return &Formatter{
		css: DefaultCSS,
	}
}

func (f *Formatter) SetCSS(css string) {
	f.css = css
}

func (f *Formatter) GetCSS() string {
	return f.css
}
