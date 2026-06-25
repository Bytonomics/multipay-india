import React from "react";

export interface RazorpayLogoProps extends React.SVGAttributes<SVGSVGElement> {
  className?: string;
}

export const RazorpayLogo: React.FC<RazorpayLogoProps> = ({
  className = "",
  ...props
}) => {
  return (
    <svg
      className={className}
      viewBox="0 0 200 50"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="Razorpay"
      {...props}
    >
      {/* Razorpay Logo - Icon + Text */}
      {/* Razorpay icon/brand mark */}
      <g>
        {/* Main shape - stylized "R" */}
        <path
          d="M15 10 L35 10 L35 15 L25 15 L25 20 L32 20 L32 25 L25 25 L25 35 L20 35 L20 20 L15 20 Z"
          fill="#3395ff"
        />
        {/* Accent element */}
        <rect x="28" y="28" width="8" height="8" rx="1" fill="#072649" />
      </g>

      {/* Razorpay text */}
      <text
        x="45"
        y="35"
        fontFamily="Arial, sans-serif"
        fontSize="24"
        fontWeight="bold"
        fill="#072649"
      >
        Razorpay
      </text>
    </svg>
  );
};
