var React = require('react/addons');

var Dropzone = React.createClass({
  getInitialState: function() {
    return {
      isDragActive: false
    }
  },

  handleReset: function (e) {
    e.preventDefault();
    this.replaceState(this.getInitialState());
  },

  handleClick: function (e) {
    // emulate clicking to upload a file
    e.preventDefault();
    var fileField = document.getElementById('fileField'),
        that = this;
    fileField.click();
    fileField.onchange = function (e) {
      var file = fileField.files[0];
      that.handleUpload(file);
    };
  },

  handleDragLeave: function(e) {
    this.setState({
      isDragActive: false
    });
  },

  handleDragOver: function(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = "copy";

    this.setState({
      isDragActive: true
    });
  },

  handleDrop: function(e) {
    e.preventDefault();

    this.setState({isDragActive: false});

    var file = e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files[0];
    this.handleUpload(file);
  },

  changeFilename: function(original, append) {
    var root = original.split('.').slice(0,-1).join(''),
        ext = original.split('.').splice(-1);
    return root + append + '.' + ext;
  },

  handleUpload: function(file) {
    var formData = new FormData(document.forms[0]),
        xhr = new XMLHttpRequest(),
        fileReader = new FileReader(),
        that = this;

    this.setState({
      isDragActive: false,
      isUploading: true
    });

    // read image to preview upload in browser
    fileReader.readAsDataURL(file);
    fileReader.onload = function (e) {
      that.setState({
        originalFile: e.target.result
      });
    }

    // add file to form and POST
    if (!document.getElementById('fileField').value) {
      formData.append('file', file);
    }
    xhr.open('POST', '/upload/', true);
    xhr.onload = function (e) {
      that.setState({
        error: false,
        isUploading: false,
        processedFile: xhr.response,
        downloadFile: that.changeFilename(file.name, '_cleaned')
      });
    };
    xhr.onerror = function (e) {
      that.setState({
        isUploading: false,
        error: true
      });
    }
    xhr.send(formData);
  },

  render: function() {
    var classes = React.addons.classSet({
      'dropzone': true,
      'dragging': this.state.isDragActive,
      'uploading': this.state.isUploading,
      'done': this.state.processedFile,
      'error': this.state.error
    });
    var message;
    if (this.state.isDragActive) {
      message = "Drop it!"
    } else if (this.state.isUploading) {
      message = ["Uploading and processing...", <br />, "Sit tight, this will take a few seconds."];
    } else if (this.state.error) {
      message = "Something didn't work... Try again";
    } else {
      message = "Drop an image here or click to upload";
    }
    if (this.state.processedFile) {
      return (
        <div>
        <p className="download">
          <a href={this.state.processedFile} download={this.state.downloadFile}>Download your image</a>
          {' or'} <button onClick={this.handleReset}>Upload another</button>
        </p>
        <img className="processed" src={this.state.processedFile} />
        </div>
      );
    }
    return (
      <div className={classes} onDragLeave={this.handleDragLeave} onDragOver={this.handleDragOver} onDrop={this.handleDrop} onClick={this.handleClick}>
        <p>{message}</p>
        <img className="loading" src="/assets/img/loading.gif" /><br />
        <img className="original" src={this.state.originalFile} />
      </div>
    );
  }

});

React.render(
  React.createElement(Dropzone),
  document.getElementById('upload')
);
