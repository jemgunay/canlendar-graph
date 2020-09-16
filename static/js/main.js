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
            datasets: [{
                label: options.title,
                backgroundColor: 'rgb(255, 99, 132)',
                borderColor: 'rgb(255, 99, 132)',
                data: data,
                type: "scatter",
            }]
        },

        // Configuration options go here
        options: {
            title: {
                display: true,
                text: options.title
            },
            responsive: true,
            legend: {
                display: false,
            },
            scales: {
                xAxes: [{
                    scaleLabel: {
                        display: true,
                        labelString: 'Date'
                    },
                    unit: 'month',
                    type: 'time',
                    ticks: {
                        source: 'auto',
                    },
                    time: {
                        unit: 'month',
                        displayFormats: {
                            quarter: 'MMM YYYY'
                        }
                    }
                }],
                yAxes: [{
                    ticks: {
                        beginAtZero: true
                    }
                }]
            }
        }
    });
}

newGraph({
    operation: 'week-view',
    title: "Week View",
    canvasID: 'main-graph',
});
