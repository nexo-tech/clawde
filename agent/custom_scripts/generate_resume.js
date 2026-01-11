#!/usr/bin/env node

/**
 * Resume Generator Script
 * Generates a professional 1-page .docx resume using the docx library
 *
 * Requirements:
 * - npm install docx
 *
 * Usage:
 * - node generate_resume.js
 */

const docx = require('docx');
const fs = require('fs');
const path = require('path');

// Resume data structure
const resumeData = {
  name: "Test Person",
  title: "Senior Software Engineer",
  contact: {
    email: "test.person@example.com",
    phone: "(555) 123-4567",
    location: "San Francisco, CA",
    linkedin: "linkedin.com/in/testperson",
    github: "github.com/testperson"
  },
  summary: "Results-driven Senior Software Engineer with 8+ years of experience in full-stack development, cloud architecture, and team leadership. Proven track record of delivering scalable solutions and mentoring junior developers.",
  experience: [
    {
      title: "Senior Software Engineer",
      company: "Tech Innovations Inc.",
      location: "San Francisco, CA",
      dates: "Jan 2021 - Present",
      bullets: [
        "Led development of microservices architecture serving 2M+ users, reducing latency by 40%",
        "Mentored team of 5 engineers and established code review practices improving quality by 30%",
        "Architected CI/CD pipeline using GitHub Actions, reducing deployment time from 2hrs to 15min"
      ]
    },
    {
      title: "Software Engineer",
      company: "Digital Solutions Corp.",
      location: "San Jose, CA",
      dates: "Jun 2018 - Dec 2020",
      bullets: [
        "Developed React-based dashboard managing $50M+ in transactions with 99.9% uptime",
        "Implemented RESTful APIs in Node.js handling 10K+ requests/min with sub-100ms response",
        "Collaborated with product team to deliver 15+ features across agile sprint cycles"
      ]
    },
    {
      title: "Junior Developer",
      company: "StartupXYZ",
      location: "Mountain View, CA",
      dates: "Aug 2016 - May 2018",
      bullets: [
        "Built responsive web applications using React, Redux, and TypeScript for 50K+ users",
        "Optimized database queries reducing load times by 60% and improving user experience"
      ]
    }
  ],
  education: {
    degree: "Bachelor of Science in Computer Science",
    school: "University of California, Berkeley",
    year: "2016"
  },
  skills: "JavaScript, TypeScript, React, Node.js, Python, Go, AWS, Docker, Kubernetes, PostgreSQL, MongoDB, Redis, Git, CI/CD, Microservices, REST APIs, GraphQL, Agile/Scrum"
};

// Create document with proper margins and styling
const { Document, Packer, Paragraph, TextRun, HeadingLevel, AlignmentType, convertInchesToTwip } = docx;

function createResume(data) {
  const doc = new Document({
    sections: [{
      properties: {
        page: {
          margin: {
            top: convertInchesToTwip(0.5),
            right: convertInchesToTwip(0.5),
            bottom: convertInchesToTwip(0.5),
            left: convertInchesToTwip(0.5),
          },
        },
      },
      children: [
        // Name (24pt)
        new Paragraph({
          text: data.name,
          heading: HeadingLevel.TITLE,
          alignment: AlignmentType.CENTER,
          spacing: { after: 100 },
          children: [
            new TextRun({
              text: data.name,
              size: 48, // 24pt = 48 half-points
              bold: true,
            }),
          ],
        }),

        // Contact Info (10pt)
        new Paragraph({
          alignment: AlignmentType.CENTER,
          spacing: { after: 200 },
          children: [
            new TextRun({
              text: `${data.contact.email} | ${data.contact.phone} | ${data.contact.location}`,
              size: 20,
            }),
            new TextRun({
              text: `\n${data.contact.linkedin} | ${data.contact.github}`,
              size: 20,
            }),
          ],
        }),

        // Summary Header (12pt bold)
        new Paragraph({
          children: [
            new TextRun({
              text: "PROFESSIONAL SUMMARY",
              size: 24,
              bold: true,
            }),
          ],
          spacing: { after: 100 },
          border: {
            bottom: {
              color: "000000",
              space: 1,
              style: "single",
              size: 6,
            },
          },
        }),

        // Summary (10pt)
        new Paragraph({
          text: data.summary,
          spacing: { after: 200 },
          children: [
            new TextRun({
              text: data.summary,
              size: 20,
            }),
          ],
        }),

        // Experience Header
        new Paragraph({
          children: [
            new TextRun({
              text: "PROFESSIONAL EXPERIENCE",
              size: 24,
              bold: true,
            }),
          ],
          spacing: { after: 100 },
          border: {
            bottom: {
              color: "000000",
              space: 1,
              style: "single",
              size: 6,
            },
          },
        }),

        // Experience entries
        ...data.experience.flatMap(exp => [
          new Paragraph({
            children: [
              new TextRun({
                text: exp.title,
                size: 20,
                bold: true,
              }),
              new TextRun({
                text: ` | ${exp.company}, ${exp.location}`,
                size: 20,
              }),
            ],
            spacing: { after: 50 },
          }),
          new Paragraph({
            children: [
              new TextRun({
                text: exp.dates,
                size: 20,
                italics: true,
              }),
            ],
            spacing: { after: 100 },
          }),
          ...exp.bullets.map(bullet =>
            new Paragraph({
              text: `• ${bullet}`,
              spacing: { after: 80 },
              children: [
                new TextRun({
                  text: `• ${bullet}`,
                  size: 20,
                }),
              ],
            })
          ),
        ]),

        // Education Header
        new Paragraph({
          children: [
            new TextRun({
              text: "EDUCATION",
              size: 24,
              bold: true,
            }),
          ],
          spacing: { after: 100, before: 100 },
          border: {
            bottom: {
              color: "000000",
              space: 1,
              style: "single",
              size: 6,
            },
          },
        }),

        // Education entry
        new Paragraph({
          children: [
            new TextRun({
              text: data.education.degree,
              size: 20,
              bold: true,
            }),
            new TextRun({
              text: ` | ${data.education.school}, ${data.education.year}`,
              size: 20,
            }),
          ],
          spacing: { after: 200 },
        }),

        // Skills Header
        new Paragraph({
          children: [
            new TextRun({
              text: "TECHNICAL SKILLS",
              size: 24,
              bold: true,
            }),
          ],
          spacing: { after: 100 },
          border: {
            bottom: {
              color: "000000",
              space: 1,
              style: "single",
              size: 6,
            },
          },
        }),

        // Skills
        new Paragraph({
          text: data.skills,
          children: [
            new TextRun({
              text: data.skills,
              size: 20,
            }),
          ],
        }),
      ],
    }],
  });

  return doc;
}

// Generate and save resume
async function main() {
  try {
    console.log('Generating resume for:', resumeData.name);

    const doc = createResume(resumeData);
    const buffer = await Packer.toBuffer(doc);

    const outputPath = path.join(__dirname, 'resume.docx');
    fs.writeFileSync(outputPath, buffer);

    console.log('✓ Resume generated successfully!');
    console.log('  Output:', outputPath);
    console.log('  Size:', (buffer.length / 1024).toFixed(2), 'KB');
  } catch (error) {
    console.error('Error generating resume:', error);
    process.exit(1);
  }
}

main();
