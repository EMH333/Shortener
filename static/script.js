window.addEventListener("load", function () {
    var background = this.document.getElementById("background");
    var size = {
        width: window.innerWidth || document.body.clientWidth,
        height: window.innerHeight || document.body.clientHeight
    }

    const BRICK_HEIGHT = 40;
    const BRICK_WIDTH = 110;

    for (let width = -5; width < size.width; width += BRICK_WIDTH) {
        for (let height = -5; height < size.height; height += BRICK_HEIGHT) {
            let el = document.createElement("div");
            el.style.top = height + "px";
            el.style.left = width + "px";
            el.style.backgroundColor = getRandomColor() + "CF";//not we are decresing opacity slightly
            el.className = "brick"
            background.appendChild(el);
            console.log(height)
        }
    }


    //randomize body background color
    this.document.body.style.backgroundColor = getRandomColor() +"50";//set color and oppacity
});

function getRandomColor() {
    var letters = '0123456789ABCDEF';
    var color = '#';
    for (var i = 0; i < 6; i++) {
        color += letters[Math.floor(Math.random() * 16)];
    }
    return color;
}