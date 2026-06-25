import React from "react";

export interface CashfreeLogoProps extends React.SVGAttributes<SVGSVGElement> {
  className?: string;
}

export const CashfreeLogo: React.FC<CashfreeLogoProps> = ({
  className = "",
  ...props
}) => {
  return (
    <svg
      className={className}
      viewBox="0 0 200 50"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="Cashfree"
      {...props}
    >
      {/* Cashfree Logo - Text based with accent */}
      <text
        x="10"
        y="35"
        fontFamily="Arial, sans-serif"
        fontSize="28"
        fontWeight="bold"
        fill="#072649"
      >
        Cashfree
      </text>
      <text
        x="10"
        y="35"
        fontFamily="Arial, sans-serif"
        fontSize="28"
        fontWeight="bold"
        fill="#5D9C59"
      >
        f
      </text>
      <text
        x="10"
        y="35"
        fontFamily="Arial, sans-serif"
        fontSize="28"
        fontWeight="bold"
        fill="#072649"
      >
        ree
      </text>
      {/* Accent dot */}
      <circle cx="145" cy="20" r="4" fill="#5D9C59" />
    </svg>
  );
};
