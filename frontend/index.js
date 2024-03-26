function getUsersCount() {
    fetch("http://localhost:31743/users-count").then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    }).then(data => {
        document.getElementById('userCount').textContent = data;
    }).catch(error => {
        console.error('There was a problem with the fetch operation:', error);
    });
}
var intervalId = setInterval(getUsersCount, 1000)