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

            // draw graph
            drawGraph(options, data);
        }
    });
}

function drawGraph(options, data) {
    let ctx = document.getElementById(options.canvasID).getContext('2d');
    let chart = new Chart(ctx, {
        // The type of chart we want to create
        type: 'line',

        // The data for our dataset
        data: {
            datasets: [
                // units consumed line
                {
                    data: data,
                    label: 'Units Consumed',
                    borderColor: 'rgb(255, 99, 132)',
                    backgroundColor: 'rgba(255, 99, 132, 0.1)',
                    borderWidth: 2,
                    lineTension: 0
                },
                // recommended units
                {
                    data: [
                        {y: 14, t: data[0].t},
                        {y: 14, t: data[data.length - 1].t}
                    ],
                    label: 'Recommended Units',
                    borderColor: 'rgb(109,109,109)',
                    fill: false,
                    borderWidth: 2,
                    borderDash: [10]
                }
            ]
        },

        // Configuration options go here
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
                        labelString: 'Week'
                    },
                    type: 'time',
                    ticks: {
                        source: 'auto',
                    },
                    time: {
                        unit: 'week',
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
    });
}

$('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
    console.log(e.target)

    if ($(e.target).attr('id') === 'week-tab') {
        newGraph({
            view: 'week',
            title: "Units Consumed per Week",
            canvasID: 'main-graph',
        });
    } else {
        newGraph({
            view: 'month',
            title: "Units Consumed per Month",
            canvasID: 'main-graph',
        });
    }
})

newGraph({
    view: 'week',
    title: "Units Consumed per Week",
    canvasID: 'main-graph',
});