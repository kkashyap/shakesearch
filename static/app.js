const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results, data.query);
      });
    });
  },

  updateTable: (results, query) => {
    const table = document.getElementById("table-body");
    const rows = [];
    let queryRegEx = new RegExp(query, 'ig');
    let resultString = results.length == 1 ? "result" : "results";
    rows.push(`<th>Found ${results.length} ${resultString} for ${query}</th>`);
    for (let result of results) {
      let searchResult = result.ResultString.replace(queryRegEx, function(str) {return '<b>'+str+'</b>'});
      rows.push(`<tr>${searchResult}... <i>(Found in ${result.WorkTitle})</i><tr/><hr class="solid"></hr>`);
    }
    table.innerHTML = rows.join("")
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
