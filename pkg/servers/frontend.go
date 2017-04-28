package servers

import "html/template"

const (
	indexHTML = `<!DOCTYPE html>
<html>
	<head>
		<title>Welcome to my service</title>
		<style type="text/css">
			#footer {
				border-top: 10px solid #005ea5;
			    background-color: #dee0e2;
			}
			#footer ul {
				list-style: none;
			}
			#footer ul li {
    			display: inline-block;
    			margin: 0 15px 15px 0;
			}
			#overview p {
				margin: 0 25px 0 25px;
			}
			.floated-inner-block {
				margin: 0 25px;
			}
			.homepage-top {
    			background: #005ea5;
    			color: #fff;
			}
			.homepage-top h1 {
				font-family: Arial, sans-serif;
    			font-size: 32px;
    			line-height: 1.09375;
    			text-transform: none;
    			font-size-adjust: 0.5;
    			font-weight: bold;
    			padding: 25px 0 15px;
			}
			.values-list ul {
				list-style: none;
    			padding: 0 25px;
			}
			.visuallyhidden {
 			   position: absolute;
    			left: -9999em;
			}
			p {
				font-family: Arial, sans-serif;
    			font-size: 16px;
				line-height: 1.25;
    			font-weight: 400;
    			text-transform: none;
			}
		</style>
	</head>
	<body>
		<header class="homepage-top">
			<div class="floated-inner-block">
				<h1>Welcome!</h1>
				<p>A simple app using for examining telemetry options.</p>
			</div>
		</header>
		<main>
			<section id="overview" aria-labelledby="overview-label">
				<h2 id="overview-label" class="visuallyhidden">Overview</h2>
				<p>This is a toy application which makes calls to upstream services.</p>
				<p>The upstream services might fail, or take a while to respond. This gives us "interesting" data to capture and then report on.</p>
			</section>
			<section id="responses" aria-labelledby="responses-label">
				<h2 id="responses-label" class="visuallyhidden">Responses</h2>
				<div class="values-list">
					<ul>
					{{range .}}
						<li>
							<code>{{.Key}}</code> : {{.Value}}
						</li>
					{{end}}
					</ul>
				</div>
			</section>
		</main>
		<footer id="footer">
			<div class="footer-meta">
				<h2 class="visuallyhidden">Support links</h2>
				<ul>
					<li><a href="https://github.com/jabley/monitoring-spike">Source</a></li>
					<li>Built by <a href="https://twitter.com/jabley">James Abley</a></li>
				</ul>
			</div>
		</footer>
	</body>
</html>
`
)

var (
	tmpl = template.Must(template.New("index.html").Parse(indexHTML))
)
