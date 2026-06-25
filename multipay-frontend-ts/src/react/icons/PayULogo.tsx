import React from "react";

export interface PayULogoProps extends React.SVGAttributes<SVGSVGElement> {
  className?: string;
}

export const PayULogo: React.FC<PayULogoProps> = ({
  className = "",
  ...props
}) => {
  return (
    <svg
      className={className}
      viewBox="0 0 200 50"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="PayU"
      {...props}
    >
      {/* PayU Logo - Text based */}
      <text
        x="10"
        y="35"
        fontFamily="Arial, sans-serif"
        fontSize="28"
        fontWeight="bold"
        fill="#EB6804"
      >
        Pay
      </text>
      <text
        x="58"
        y="35"
        fontFamily="Arial, sans-serif"
        fontSize="28"
        fontWeight="bold"
        fill="#072649"
      >
        U
      </text>
      {/* Accent line */}
      <rect x="10" y="40" width="60" height="3" rx="1" fill="#EB6804" />
    </svg>
  );
};
