function newGraph(options) {
    // fetch graph data from server
    $.ajax({
        url: '/data?view=' + options.view,
        type: 'GET',
        dataType: 'json',
        error: function (e) {
            alert('failed to retrieve ' + options.view + ' data (' + e.status + ')')
        },
        success: function (data) {
            console.log(data);
            drawGraph(options, data);
        }
    });
}

let currentChart = null

function drawGraph(options, data) {
    // clean up previous chart
    if (currentChart !== null) {
        currentChart.destroy();
    }

    let chartConfig = {
        type: options.type,

        // the data for our dataset
        data: {
            datasets: [{
                // units consumed line
                data: data.plots,
                label: 'Units Consumed',
                borderColor: 'rgb(255, 99, 132)',
                backgroundColor: 'rgba(255, 99, 132, 0.1)',
                borderWidth: 2,
                lineTension: 0
            }]
        },

        // configuration options go here
        options: {
            title: {
                display: true,
                text: options.title
            },
            responsive: true,
            scales: {
                xAxes: [{
                    scaleLabel: {
                        display: true,
                        labelString: options.view[0].toUpperCase() + options.view.substring(1)
                    },
                    type: 'time',
                    ticks: {
                        source: 'auto',
                    },
                    time: {
                        unit: options.unit,
                        isoWeekday: true,
                        displayFormats: {
                            quarter: 'MMM YYYY'
                        }
                    }
                }],
                yAxes: [{
                    scaleLabel: {
                        display: true,
                        labelString: 'Units'
                    },
                    ticks: {
                        beginAtZero: true
                    }
                }]
            },
            legend: {
                display: true,
                labels: {
                    fontColor: 'rgb(255, 99, 132)'
                },
                position: 'right'
            }
        }
    };

    // display UK units guideline
    if (options.enableGuideline === true) {
        chartConfig.data.datasets.push({
            data: [
                {y: data.config.guideline, t: data.plots[0].t},
                {y: data.config.guideline, t: data.plots[data.plots.length - 1].t}
            ],
            label: 'UK Guideline Units',
            borderColor: 'rgb(109,109,109)',
            fill: false,
            borderWidth: 2,
            borderDash: [10]
        });
    }

    // create new chart
    let ctx = document.getElementById('main-graph').getContext('2d');
    currentChart = new Chart(ctx, chartConfig)
}

function selectGraph(view) {
    switch (view) {
        case 'month':
            newGraph({
                view: 'month',
                type: 'line',
                title: 'Units Consumed per Month',
                enableGuideline: true
            });
            break;
        case 'week':
            newGraph({
                view: 'week',
                type: 'line',
                title: 'Units Consumed per Week',
                enableGuideline: true
            });
            break;
        case 'day':
            newGraph({
                view: 'day',
                type: 'scatter',
                title: 'Units Consumed per Day',
            });
            break;
    }
}

// listen for nav bar tab switch to trigger graph creation
$('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
    let id = $(e.target).attr('id');
    selectGraph(id.substring(0, id.length - 4));
})

// draw initial graph
selectGraph('month')