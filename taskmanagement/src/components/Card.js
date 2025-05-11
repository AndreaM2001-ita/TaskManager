import React from 'react';
import './Card.css'; // Import the CSS file

const Card = ({ children }) => {
  return (
    <div className="card">
      {children}
    </div>
  );
};

export default Card;
