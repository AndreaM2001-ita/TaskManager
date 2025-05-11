import React from 'react';
import './Button.css';  // Import the CSS file

const Button = ({ onClick, children, type = "button" }) => {
  return (
    <button
      type={type}
      onClick={onClick}
      className="btn"
    >
      {children}
    </button>
  );
};

export default Button;