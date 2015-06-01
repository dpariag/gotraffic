function TxRxGraph() {
	var seriesData = [
		[], // Tx
		[]  // Rx
	];
	var palette = new Rickshaw.Color.Palette({scheme: 'cool'});
	var txcolor = palette.color();
	palette.color(); // skip this color
	var rxcolor = palette.color();
	var series = [
		{
			color: txcolor,
			data: seriesData[0],
			name: 'Tx'
		}, {
			color: rxcolor,
			data: seriesData[1],
			name: 'Rx'
		}
	];
	var graph = new Rickshaw.Graph({
		element: document.getElementById('graph'),
		height: 150,
		renderer: 'line',
		stroke: true,
		preserve: true,
		unstack: true,
		series: series
	});
	graph.render();
	var legend = new Rickshaw.Graph.Legend({
		graph: graph,
		element: document.getElementById('legend')
	});
	var shelving = new Rickshaw.Graph.Behavior.Series.Toggle({
		graph: graph,
		legend: legend
	});
	var order = new Rickshaw.Graph.Behavior.Series.Order({
		graph: graph,
		legend: legend
	});
	var highlighter = new Rickshaw.Graph.Behavior.Series.Highlight({
		graph: graph,
		legend: legend
	});
	var ticksTreatment = 'glow';
	var xAxis = new Rickshaw.Graph.Axis.Time({
		graph: graph,
		ticksTreatment: ticksTreatment,
		timeFixture: new Rickshaw.Fixtures.Time.Local(),
	});
	xAxis.render();
	var yAxis = new Rickshaw.Graph.Axis.Y({
		graph: graph,
		tickFormat: Rickshaw.Fixtures.Number.formatKMBT,
		ticksTreatment: ticksTreatment
	});
	yAxis.render();
	var hoverDetail = new Rickshaw.Graph.HoverDetail({
		graph: graph,
	    	xFormatter: function(x) {
			return new Date(x * 1000).toString();
		}
	});
	return {graph: graph, series: seriesData};
}

function Counter() {
	this.last = 0;
	this.addPoint = function(point) {
		if (this.last == 0) {
			this.last = point;
			return 0;
		}
		var v = point - this.last;
		this.last = point;
		if (v < 0) v = 0;
		return v;
	}
}

$(document).ready(function() {
	var txrx = TxRxGraph();
	var tx = new Counter();
	var rx = new Counter();
	var ev = new EventSource('./sse');
	ev.onmessage = function(event) {
		var d = JSON.parse(event.data);
		txrx.series[0].push({'x': d.time, 'y': tx.addPoint(d.tx)});
		txrx.series[1].push({'x': d.time, 'y': rx.addPoint(d.rx)});
		// max of 20 data points.
		if (txrx.series[0].length == 20) {
			txrx.series[0].splice(0, 1);
			txrx.series[1].splice(0, 1);
		}
		txrx.graph.update();
	};
	window.addEventListener('resize', function() {
		txrx.graph.configure({
			width: document.getElementById('container').offsetWidth * 0.95
		});
		txrx.graph.render();
	});
});
