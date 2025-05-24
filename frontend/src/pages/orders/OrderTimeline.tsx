import React from 'react';
import './OrderTimeline.css';
import { TimelineEvent } from '../../api/Order';

interface OrderTimelineProps {
  status: 'Created' | 'Confirmed' | 'Delivered' | 'Cancelled' | 'Processing' | 'Payment Failed';
  timeline: TimelineEvent[];
}

const OrderTimeline: React.FC<OrderTimelineProps> = ({ status, timeline }) => {
  const sortedTimeline = [...timeline].sort((a, b) => 
    new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  return (
    <div className="order-timeline">
      {sortedTimeline.map((event, index) => {
        const isCreated = event.name === "Created";
        const isFailed = event.name === "Payment Failed";
        const isCancelled = event.name === "Cancelled";
        const isDelivered = event.name === "Delivered";
        const isPaymentCompleted = event.name === "Payment Completed";
        const isProcessing = event.name === "Processing Payment";
        const isPaymentFailed = event.name === "Payment Failed";
        const isCurrent = event.name === status || (status === 'Processing' && isProcessing);
        return (
          <div key={index} className="timeline-step">
            <div className="timeline-marker-container">
              <div className={`timeline-marker 
                ${isCurrent ? 'current' : ''} 
                ${isCreated ? 'created' : ''} 
                ${isFailed ? 'failed' : ''} 
                ${isCancelled ? 'cancelled' : ''} 
                ${isDelivered ? 'delivered' : ''}
                ${isProcessing ? 'processing' : ''}`}>
                {isProcessing && <div className="processing-dot"></div>}
              </div>
              {index < sortedTimeline.length - 1 && (
                <div className={`timeline-connector 
                  ${isCreated ? 'created' : ''} 
                  ${isFailed ? 'failed' : ''} 
                  ${isCancelled ? 'cancelled' : ''} 
                  ${isDelivered ? 'delivered' : ''}
                  ${isProcessing ? 'processing' : ''}`}>
                </div>
              )}
            </div>
            <div className="timeline-content">
              <p className={`timeline-label 
                ${isCreated ? 'created' : ''} 
                ${isFailed || isPaymentFailed ? 'failed' : ''} 
                ${isCancelled ? 'cancelled' : ''} 
                ${isDelivered ? 'delivered' : ''} 
                ${isPaymentCompleted ? 'payment-completed' : ''}
                ${isProcessing ? 'processing' : ''}`}>
                {event.name === 'Processing Payment' ? 'Processing' : event.name}
              </p>
              <p className="timeline-date">{new Date(event.timestamp).toLocaleString()}</p>
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default OrderTimeline;
