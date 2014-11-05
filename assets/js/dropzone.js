/**
 * @jsx React.DOM
 */

var React = require('react/addons');

var Dropzone = React.createClass({
  getInitialState: function() {
    return {
      isDragActive: false
    }
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

    this.setState({
      isDragActive: false
    });

    if (this.props.handler) {
      var file = e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files[0];
      this.props.handler(file);
    } else {
      console.error('No handler specified to accept the dropped file.');
    }
  },
  render: function() {
    var classes = React.addons.classSet({
      'dropzone': true,
      'dragging': this.state.isDragActive
    });
    return (
      <div className={classes} onDragLeave={this.handleDragLeave} onDragOver={this.handleDragOver} onDrop={this.handleDrop}>
        {this.props.children || <span>{this.props.message || "Drop Here"}</span>}
      </div>
    );
  }

});

module.exports = Dropzone;
