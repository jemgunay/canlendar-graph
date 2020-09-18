function newGraph(options) {
    // fetch graph data from server
    $.ajax({
        url: '/data?operation=' + options.operation,
        type: 'GET',
        dataType: 'json',
        error: function (e) {
            alert('failed to retrieve ' + options.operation + ' data (' + e.status + ')')
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
                    lineTension: 0,
                },
                // recommended units
                {
                    data: [
                        {y: 14, t: data[0].t},
                        {y: 14, t: data[data.length - 1].t}
                    ],
                    label: 'Recommended Units',
                    borderColor: 'rgb(81,81,81)',
                    fill: false
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

newGraph({
    operation: 'week-view',
    title: "Units Consumed per Week",
    canvasID: 'main-graph',
});
